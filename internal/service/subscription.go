package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/infra"
	"github.com/iokiris/efm-subscription-api/internal/logger"
	"github.com/iokiris/efm-subscription-api/internal/model"
	"github.com/iokiris/efm-subscription-api/internal/repo"

	"go.uber.org/zap"
)

type SubscriptionService struct {
	repo      repo.SubscriptionRepoInterface
	redis     RedisInterface
	publisher Publisher
	ttl       time.Duration
	metrics   *infra.Metrics
}

func NewSubscriptionService(r repo.SubscriptionRepoInterface, redisClient RedisInterface, pub Publisher, ttl time.Duration) *SubscriptionService {
	return &SubscriptionService{
		repo:      r,
		redis:     redisClient,
		publisher: pub,
		ttl:       ttl,
		metrics:   nil, // будет установлено через SetMetrics
	}
}

// SetMetrics устанавливает метрики для сервиса
func (s *SubscriptionService) SetMetrics(metrics *infra.Metrics) {
	s.metrics = metrics
}

// -------------------- CRUD --------------------

func (s *SubscriptionService) Create(ctx context.Context, sub *model.Subscription) error {
	if err := s.repo.Create(ctx, sub); err != nil {
		logger.L.Error("subscription.create.failed", zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, sub.UserID)
	s.publishEvent("subscriptions", "created", sub)

	// Метрики
	if s.metrics != nil {
		s.metrics.SubscriptionsCreated.Inc()
		s.metrics.SubscriptionsTotal.WithLabelValues(sub.Service, "active").Inc()
	}

	logger.L.Info("subscription.create.ok",
		zap.Int64("id", sub.ID),
		zap.String("user_id", sub.UserID),
		zap.String("service", sub.Service),
	)
	return nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub *model.Subscription) error {
	if err := s.repo.Update(ctx, sub); err != nil {
		logger.L.Error("subscription.update.failed", zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, sub.UserID)
	s.publishEvent("subscriptions", "updated", sub)

	// Метрики
	if s.metrics != nil {
		s.metrics.SubscriptionsUpdated.Inc()
	}

	logger.L.Info("subscription.update.ok", zap.Int64("id", sub.ID))
	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id int64, userID string) error {
	// Если userID не передан, попробуем получить из БД
	if userID == "" {
		sub, err := s.repo.GetByID(ctx, id)
		if err == nil && sub != nil {
			userID = sub.UserID
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.L.Error("subscription.delete.failed", zap.Error(err))
		return err
	}

	s.invalidateCache(ctx, userID)
	s.publishEvent("subscriptions", "deleted", map[string]any{"id": id, "user_id": userID})

	// Метрики
	if s.metrics != nil {
		s.metrics.SubscriptionsDeleted.Inc()
		s.metrics.SubscriptionsTotal.WithLabelValues("unknown", "deleted").Inc()
	}

	logger.L.Info("subscription.delete.ok", zap.Int64("id", id), zap.String("user_id", userID))
	return nil
}

func (s *SubscriptionService) Get(ctx context.Context, id int64) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.L.Error("subscription.get.failed", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}
	return sub, nil
}

// List возвращает список подписок пользователя
func (s *SubscriptionService) List(ctx context.Context, userID string) ([]model.Subscription, error) {
	subs, err := s.repo.List(ctx, userID)
	if err != nil {
		logger.L.Error("subscription.list.failed", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	return subs, nil
}

// -------------------- Summary --------------------

// GetSummary принимает строки from/to, парсит их в time.Time и вызывает repo.GetSummary.
// Пустые from/to — означают "всё" Формат даты: 01-2005.
func (s *SubscriptionService) GetSummary(ctx context.Context, userID string, serviceName string, from, to string) (int, error) {
	key := fmt.Sprintf("summary:%s:%s:%s-%s", userID, serviceName, from, to)

	// кеш — если есть, вернуть
	if s.redis != nil {
		if val, err := s.redis.Get(ctx, key).Result(); err == nil {
			var total int
			if _, err := fmt.Sscanf(val, "%d", &total); err == nil {
				logger.L.Debug("summary.cache.hit",
					zap.String("user_id", userID),
					zap.String("service", serviceName),
				)
				return total, nil
			}
		}
	}

	// парсим даты
	fromT, toT, err := normalizeRangeMY(from, to)
	logger.L.Debug("summary.range",
		zap.String("fromMY", fromT.String()),
		zap.String("toMY", toT.String()))

	if err != nil {
		logger.L.Error("summary.parse_dates.failed",
			zap.String("from", from), zap.String("to", to), zap.Error(err))
		return 0, err
	}

	total, err := s.repo.GetSummary(ctx, userID, serviceName, fromT, toT)
	if err != nil {
		logger.L.Error("summary.query.failed", zap.Error(err))
		return 0, err
	}

	// записать в кеш (если есть)
	if s.redis != nil {
		if err := s.redis.Set(ctx, key, total, s.ttl).Err(); err != nil {
			logger.L.Warn("summary.cache.set_failed", zap.String("key", key), zap.Error(err))
		}
	}

	// Метрики
	if s.metrics != nil {
		s.metrics.SubscriptionsSummary.WithLabelValues(serviceName).Observe(float64(total))
	}

	logger.L.Info("summary.ok",
		zap.String("user_id", userID),
		zap.String("service", serviceName),
		zap.Int("total", total),
	)
	return total, nil
}

// -------------------- Helpers --------------------

func (s *SubscriptionService) invalidateCache(ctx context.Context, userID string) {
	if s.redis == nil {
		return
	}
	pattern := fmt.Sprintf("summary:%s:*", userID)
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.redis.Del(ctx, iter.Val()).Err(); err != nil {
			logger.L.Warn("cache.del.failed",
				zap.String("key", iter.Val()),
				zap.Error(err))
		}
	}
}

func (s *SubscriptionService) publishEvent(exchange, routingKey string, payload interface{}) {
	if s.publisher == nil {
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.L.Error("publish.marshal.failed", zap.Error(err))
		return
	}

	if err := s.publisher.Publish(exchange, routingKey, data); err != nil {
		logger.L.Error("publish.failed",
			zap.String("exchange", exchange),
			zap.String("routing_key", routingKey),
			zap.Error(err))
	}
}

// normalizeRangeMY парсит строки в time.Time
func normalizeRangeMY(from, to string) (time.Time, time.Time, error) {
	var fromMY, toMY model.MonthYear
	var err error

	if from == "" {
		fromMY = model.MonthYear(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)) // минимальная дата
	} else {
		fromMY, err = parseMonthYear(from)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if to == "" {
		toMY = model.MonthYear(time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)) // far future, первый день месяца
	} else {
		toMY, err = parseMonthYear(to)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return time.Time(fromMY), time.Time(toMY), nil
}

func parseMonthYear(s string) (model.MonthYear, error) {
	var my model.MonthYear
	if err := my.UnmarshalJSON([]byte(`"` + s + `"`)); err != nil {
		return model.MonthYear{}, err
	}
	return my, nil
}
