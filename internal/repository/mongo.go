package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LinkStorageService/internal/domain"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	collection := client.Database(dbName).Collection("links")

	repo := &MongoRepository{
		collection: collection,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := repo.ensureIndexes(ctx); err != nil {
		log.Printf("WARNING: failed to create indexes: %v", err)
	} else {
		log.Println("MongoDB indexes created successfully")
	}

	return repo
}

func (r *MongoRepository) ensureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "short_code", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_short_code"),
	})
	if err != nil {
		return err
	}

	_, err = r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: -1}},
		Options: options.Index().SetName("idx_created_at"),
	})
	if err != nil {
		return err
	}

	return err
}

func (r *MongoRepository) Create(ctx context.Context, link *domain.Link) error {
	_, err := r.collection.InsertOne(ctx, link)
	return err
}

func (r *MongoRepository) FindByCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	var link domain.Link

	filter := bson.M{"short_code": shortCode}
	err := r.collection.FindOne(ctx, filter).Decode(&link)

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("link not found")
	}

	return &link, err
}

// IncrementAndGetVisits атомарно увеличивает счётчик и возвращает обновлённую сущность
func (r *MongoRepository) IncrementAndGetVisits(ctx context.Context, shortCode string) (*domain.Link, error) {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$inc": bson.M{"visits": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var link domain.Link
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&link)

	if err == mongo.ErrNoDocuments {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *MongoRepository) IncrementVisitsOnly(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}
	update := bson.M{"$inc": bson.M{"visits": 1}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("link not found")
	}

	return nil
}

func (r *MongoRepository) List(ctx context.Context, limit, offset int) ([]*domain.Link, int64, error) {
	total, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var links []*domain.Link
	for cursor.Next(ctx) {
		var link domain.Link
		if err := cursor.Decode(&link); err != nil {
			return nil, 0, err
		}
		links = append(links, &link)
	}

	return links, total, nil
}

func (r *MongoRepository) Delete(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}
	result, err := r.collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("link not found")
	}

	return nil
}

func (r *MongoRepository) Exists(ctx context.Context, shortCode string) (bool, error) {
	filter := bson.M{"short_code": shortCode}
	count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	return count > 0, err
}
