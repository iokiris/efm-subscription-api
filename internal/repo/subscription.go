package repo

import (
	"context"
	"time"

	"github.com/iokiris/efm-subscription-api/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionRepoInterface интерфейс для репозитория подписок
type SubscriptionRepoInterface interface {
	GetByID(ctx context.Context, id int64) (*model.Subscription, error)
	Create(ctx context.Context, s *model.Subscription) error
	Update(ctx context.Context, s *model.Subscription) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, userID string) ([]model.Subscription, error)
	GetSummary(ctx context.Context, userID, service string, from, to time.Time) (int, error)
}

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(db *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id int64) (*model.Subscription, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`
	var s model.Subscription
	err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Service, &s.Price, &s.UserID,
		&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubscriptionRepo) Create(ctx context.Context, s *model.Subscription) error {
	const q = `
        INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `
	return r.db.QueryRow(ctx, q,
		s.Service, s.Price, s.UserID, s.StartDate, s.EndDate,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SubscriptionRepo) Update(ctx context.Context, s *model.Subscription) error {
	const q = `
        UPDATE subscriptions
        SET service_name=$1, price=$2, start_date=$3, end_date=$4, updated_at=NOW()
        WHERE id=$5
        RETURNING updated_at
    `
	return r.db.QueryRow(ctx, q,
		s.Service, s.Price, s.StartDate, s.EndDate, s.ID,
	).Scan(&s.UpdatedAt)
}

func (r *SubscriptionRepo) Delete(ctx context.Context, id int64) error {
	ct, err := r.db.Exec(ctx, "DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SubscriptionRepo) List(ctx context.Context, userID string) ([]model.Subscription, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE user_id=$1
        ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(
			&s.ID, &s.Service, &s.Price, &s.UserID,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// GetSummary возвращает сумму цен по подпискам пользователя за указанный период.
// Даты from/to могут быть в формате YYYY-MM (например, "05-2025") или RFC3339.
// Пустые from/to означают весь период.
// Подписки учитываются только если они пересекают интервал [from; to]:
// NOTE: кеширование через Redis по ключу userID:service:from-to.
func (r *SubscriptionRepo) GetSummary(ctx context.Context, userID, service string, from, to time.Time) (int, error) {
	const q = `SELECT COALESCE(SUM(price), 0)
	FROM subscriptions
	WHERE user_id = $1
	  AND ($2 = '' OR service_name = $2)
	  AND start_date <= $4
	  AND (end_date IS NULL AND start_date <= $4   -- активная подписка пересекает интервал
		   OR end_date >= $3)                      -- завершённая подписка пересекает интервал
	`
	var total int
	err := r.db.QueryRow(ctx, q, userID, service, from, to).Scan(&total)
	return total, err
}
