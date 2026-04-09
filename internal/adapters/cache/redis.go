package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	
	"cascade/internal/application/port"
	"cascade/internal/domain"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) port.Cache {
	return &redisCache{client: client}
}

func (c *redisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}
	return val, err
}

func (c *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (c *redisCache) Increment(ctx context.Context, key string) (int64, error) {
	res, err := c.client.Incr(ctx, key).Result()
	return res, err
}

func (c *redisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return c.client.Expire(ctx, key, ttl).Err()
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
