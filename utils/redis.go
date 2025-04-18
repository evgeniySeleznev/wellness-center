package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	GetFromCache(ctx context.Context, key string) (string, error)
	SetToCache(ctx context.Context, key string, value string, expiration time.Duration) error
	Close() error
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient() (RedisClient, error) {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisClient{client: client}, nil
}

func (r *redisClient) Close() error {
	if r.client == nil {
		return nil
	}
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	return nil
}

func (r *redisClient) GetFromCache(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", errors.New("Redis client is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key '%s' not found", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get value from Redis: %w", err)
	}

	return val, nil
}

func (r *redisClient) SetToCache(ctx context.Context, key string, value string, expiration time.Duration) error {
	if r.client == nil {
		return errors.New("Redis client is not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set value in Redis: %w", err)
	}

	return nil
}
