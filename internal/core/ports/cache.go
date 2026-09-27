package ports

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, target any) (bool, error)
	Set(ctx context.Context, key string, val any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}
