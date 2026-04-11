package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"LinkStorageService/internal/domain"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RedisCache) Set(ctx context.Context, shortCode string, link *domain.Link) error {
	key := fmt.Sprintf("link:%s", shortCode)
	data, err := json.Marshal(link)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisCache) Get(ctx context.Context, shortCode string) (*domain.Link, error) {
	key := fmt.Sprintf("link:%s", shortCode)

	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var link domain.Link
	if err := json.Unmarshal(data, &link); err != nil {
		return nil, err
	}

	return &link, nil
}

func (c *RedisCache) Delete(ctx context.Context, shortCode string) error {
	key := fmt.Sprintf("link:%s", shortCode)
	return c.client.Del(ctx, key).Err()
}
