package service

import (
	"context"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/model"

	"github.com/redis/go-redis/v9"
)

// SubscriptionServiceInterface интерфейс для сервиса подписок
type SubscriptionServiceInterface interface {
	Create(ctx context.Context, sub *model.Subscription) error
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id int64, userID string) error
	Get(ctx context.Context, id int64) (*model.Subscription, error)
	List(ctx context.Context, userID string) ([]model.Subscription, error)
	GetSummary(ctx context.Context, userID, serviceName, from, to string) (int, error)
}

// RedisInterface интерфейс для Redis клиента
type RedisInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}
