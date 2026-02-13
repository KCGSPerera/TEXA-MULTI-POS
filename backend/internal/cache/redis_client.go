package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Enabled() bool
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error)
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(redisURL string) (*RedisClient, error) {
	if redisURL == "" {
		return &RedisClient{}, nil
	}

	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	return &RedisClient{client: redis.NewClient(options)}, nil
}

func (c *RedisClient) Enabled() bool {
	return c != nil && c.client != nil
}

func (c *RedisClient) Ping(ctx context.Context) error {
	if !c.Enabled() {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, bool, error) {
	if !c.Enabled() {
		return "", false, nil
	}

	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (c *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if !c.Enabled() {
		return nil
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisClient) GetOrSet(ctx context.Context, key string, ttl time.Duration, loader func() (string, error)) (string, error) {
	value, ok, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if ok {
		return value, nil
	}

	value, err = loader()
	if err != nil {
		return "", err
	}
	if err := c.Set(ctx, key, value, ttl); err != nil {
		return "", err
	}
	return value, nil
}
