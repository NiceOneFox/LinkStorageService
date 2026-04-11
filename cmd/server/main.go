package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"LinkStorageService/internal/generator"
	"LinkStorageService/internal/handler"
	"LinkStorageService/internal/repository"
	"LinkStorageService/internal/service"
)

func main() {
	// Config
	nodeID, _ := strconv.ParseInt(getEnv("NODE_ID", "1"), 10, 64)
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongoDB := getEnv("MONGO_DB", "linkstorage")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	// MongoDB connection
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}
	log.Println("MongoDB connected")

	// Redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	log.Println("Redis connected")

	// Repository & Cache
	mongoRepo := repository.NewMongoRepository(mongoClient, mongoDB)
	redisCache := repository.NewRedisCache(redisClient, 30*time.Second)

	// Generators
	snowflakeGen, err := generator.NewSnowflakeGenerator(nodeID)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	base62Encoder := generator.NewBase62Encoder()

	// Service
	linkService := service.NewLinkService(
		mongoRepo,
		redisCache,
		snowflakeGen,
		base62Encoder,
	)

	// Handler
	linkHandler := handler.NewLinkHandler(linkService)

	// Routes
	mux := http.NewServeMux()

	mux.HandleFunc("GET /links/{short_code}/stats", linkHandler.Stats)
	mux.HandleFunc("DELETE /links/{short_code}", linkHandler.Delete)
	mux.HandleFunc("GET /links/{short_code}", linkHandler.Get)
	mux.HandleFunc("GET /links", linkHandler.List)
	mux.HandleFunc("POST /links", linkHandler.Create)
	mux.HandleFunc("GET /health", linkHandler.Health)

	// Server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      withMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on :8080, NodeID=%d", nodeID)
		log.Printf("MongoDB: %s, Database: %s", mongoURI, mongoDB)
		log.Printf("Redis: %s", redisAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	if err := mongoClient.Disconnect(ctx); err != nil {
		log.Printf("MongoDB disconnect error: %v", err)
	}

	if err := redisClient.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	log.Println("Server stopped")
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		w.Header().Set("Content-Type", "application/json")

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		log.Printf("%s %s %d %v", r.Method, r.URL.Path, rw.statusCode, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
