package port

import (
	"context"
	"time"
)

// Cache abstracts the Redis operations needed for distributed flags and rate limits.
type Cache interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
	Del(ctx context.Context, key string) error

	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}
