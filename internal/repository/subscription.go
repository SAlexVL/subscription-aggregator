package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/programmerpark/subscription-aggregator/internal/model"
)

var ErrNotFound = errors.New("subscription not found")

type postgresSubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) SubscriptionRepository {
	return &postgresSubscriptionRepository{pool: pool}
}

func (r *postgresSubscriptionRepository) Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error) {
	const q = `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`

	var end *time.Time
	if req.EndDate != nil {
		t := req.EndDate.Time
		end = &t
	}

	row := r.pool.QueryRow(ctx, q, req.ServiceName, req.Price, req.UserID, req.StartDate.Time, end)
	sub, err := scanSubscription(row)
	if err != nil {
		return nil, fmt.Errorf("create subscription: %w", err)
	}
	return sub, nil
}

func (r *postgresSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	const q = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`
	sub, err := scanSubscription(r.pool.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return sub, nil
}

func (r *postgresSubscriptionRepository) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, int, error) {
	where := `WHERE user_id = $1`
	args := []any{f.UserID}
	argN := 2

	if f.ServiceName != nil && *f.ServiceName != "" {
		where += fmt.Sprintf(" AND service_name ILIKE $%d", argN)
		args = append(args, *f.ServiceName)
		argN++
	}

	var total int
	countQ := `SELECT COUNT(*) FROM subscriptions ` + where
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count subscriptions: %w", err)
	}

	q := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
	` + where + ` ORDER BY created_at DESC`

	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	result := make([]model.Subscription, 0)
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, *sub)
	}
	return result, total, rows.Err()
}

func (r *postgresSubscriptionRepository) Update(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	setParts := make([]string, 0, 6)
	args := []any{id}
	argN := 2

	if req.ServiceName != nil {
		setParts = append(setParts, fmt.Sprintf("service_name = $%d", argN))
		args = append(args, *req.ServiceName)
		argN++
	}
	if req.Price != nil {
		setParts = append(setParts, fmt.Sprintf("price = $%d", argN))
		args = append(args, *req.Price)
		argN++
	}
	if req.UserID != nil {
		setParts = append(setParts, fmt.Sprintf("user_id = $%d", argN))
		args = append(args, *req.UserID)
		argN++
	}
	if req.StartDate != nil {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", argN))
		args = append(args, req.StartDate.Time)
		argN++
	}
	if req.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argN))
		args = append(args, req.EndDate.Time)
		argN++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	q := fmt.Sprintf(`
		UPDATE subscriptions
		SET %s
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at
	`, strings.Join(setParts, ", "))

	sub, err := scanSubscription(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update subscription: %w", err)
	}
	return sub, nil
}

func (r *postgresSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Sum calculates total subscription cost for the period on the database side.
// For each matching row: price * number of overlapping months with [from, to].
func (r *postgresSubscriptionRepository) Sum(ctx context.Context, f model.SumFilter) (int, error) {
	q := `
		SELECT COALESCE(SUM(
			price * (
				(EXTRACT(YEAR FROM overlap_end)::int - EXTRACT(YEAR FROM overlap_start)::int) * 12
				+ (EXTRACT(MONTH FROM overlap_end)::int - EXTRACT(MONTH FROM overlap_start)::int)
				+ 1
			)
		), 0)::bigint
		FROM (
			SELECT
				price,
				GREATEST(start_date, $1::date) AS overlap_start,
				LEAST(COALESCE(end_date, $2::date), $2::date) AS overlap_end
			FROM subscriptions
			WHERE start_date <= $2::date
			  AND (end_date IS NULL OR end_date >= $1::date)
	`
	args := []any{f.From.Time, f.To.Time}
	argN := 3

	if f.UserID != nil {
		q += fmt.Sprintf(" AND user_id = $%d", argN)
		args = append(args, *f.UserID)
		argN++
	}
	if f.ServiceName != nil && *f.ServiceName != "" {
		q += fmt.Sprintf(" AND service_name ILIKE $%d", argN)
		args = append(args, *f.ServiceName)
	}

	q += `
		) AS overlapped
		WHERE overlap_start <= overlap_end
	`

	var total int64
	if err := r.pool.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("sum subscriptions: %w", err)
	}
	return int(total), nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanSubscription(row scannable) (*model.Subscription, error) {
	var (
		sub   model.Subscription
		start time.Time
		end   *time.Time
	)
	if err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&start,
		&end,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	); err != nil {
		return nil, err
	}

	sub.StartDate = model.YearMonth{Time: time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)}
	if end != nil {
		ym := model.YearMonth{Time: time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)}
		sub.EndDate = &ym
	}
	return &sub, nil
}
