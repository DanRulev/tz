package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"tz/internal/domain"

	"github.com/google/uuid"
)

type Query interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type SubscriptionRepository struct {
	db Query
}

func NewSubscriptionRepository(db Query) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (s *SubscriptionRepository) CreateSubscription(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	query := `INSERT INTO subscriptions (id, user_id, service_name, start_date, end_date, price) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, user_id, service_name, start_date, end_date, price, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query, sub.ID, sub.UserID, sub.ServiceName, sub.StartDate, sub.EndDate, sub.Price).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.ServiceName,
		&sub.StartDate,
		&sub.EndDate,
		&sub.Price,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if err != nil {
		return domain.Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionRepository) SubscriptionByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error) {
	query := `
		SELECT
			id,
			user_id,
			service_name,
			start_date,
			end_date,
			price,
			created_at,
			updated_at
		FROM subscriptions
		WHERE id = $1
	`

	var sub domain.Subscription
	err := s.db.GetContext(ctx, &sub, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, domain.ErrNotFound
		}
		return domain.Subscription{}, fmt.Errorf("failed to get subscription: %w", err)
	}

	return sub, nil
}

func (s *SubscriptionRepository) Subscriptions(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, int, error) {
	var (
		where  []string
		args   []interface{}
		argIdx = 1
	)

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.ServiceName != nil {
		where = append(where, fmt.Sprintf("service_name = $%d", argIdx))
		args = append(args, *filter.ServiceName)
		argIdx++
	}

	var whereClause string
	if len(where) > 0 {
		whereClause = fmt.Sprintf("WHERE %v", strings.Join(where, " AND "))
	}

	var total int
	totalQuery := "SELECT COUNT(*) FROM subscriptions " + whereClause
	err := s.db.GetContext(ctx, &total, totalQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count of subscriptions: %w", err)
	}

	if total == 0 {
		return []domain.Subscription{}, 0, nil
	}

	args = append(args, filter.Limit, filter.Offset)

	query := fmt.Sprintf(`
    SELECT 
		id,
		user_id,
		service_name,
		price,
		start_date,
		end_date,
		created_at,
		updated_at
	FROM subscriptions
	%s
	LIMIT $%d OFFSET $%d
    `, whereClause, argIdx, argIdx+1)

	var subs []domain.Subscription
	err = s.db.SelectContext(ctx, &subs, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subs, total, err
}

func (s *SubscriptionRepository) SubscriptionsCost(ctx context.Context, filter domain.CostRequest) ([]domain.Subscription, error) {
	var (
		where  []string
		args   []interface{}
		argIdx = 1
	)

	if filter.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.ServiceName != nil {
		where = append(where, fmt.Sprintf("service_name = $%d", argIdx))
		args = append(args, *filter.ServiceName)
		argIdx++
	}

	var whereClause string
	if len(where) > 0 {
		whereClause = fmt.Sprintf("WHERE %v", strings.Join(where, " AND "))
	}

	query := `SELECT id, user_id, service_name, price, start_date, end_date, created_at, updated_at
		FROM subscriptions ` + whereClause

	var subs []domain.Subscription
	err := s.db.SelectContext(ctx, &subs, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionRepository) UpdateSubscription(ctx context.Context, id uuid.UUID, sub domain.UpdateSubscription) (domain.Subscription, error) {
	var (
		set    []string
		args   []interface{}
		argIdx = 1
	)

	if sub.ServiceName != nil {
		set = append(set, fmt.Sprintf("service_name = $%d", argIdx))
		args = append(args, *sub.ServiceName)
		argIdx++
	}
	if sub.Price != nil {
		set = append(set, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, *sub.Price)
		argIdx++
	}
	if sub.StartDate != nil {
		set = append(set, fmt.Sprintf("start_date = $%d", argIdx))
		args = append(args, *sub.StartDate)
		argIdx++
	}
	if sub.EndDate != nil {
		set = append(set, fmt.Sprintf("end_date = $%d", argIdx))
		args = append(args, *sub.EndDate)
		argIdx++
	}

	if len(set) == 0 {
		return s.SubscriptionByID(ctx, id)
	}

	set = append(set, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE subscriptions SET %s WHERE id = $%d RETURNING id, user_id, service_name, price, start_date, end_date, created_at, updated_at",
		strings.Join(set, ", "),
		argIdx,
	)

	var subscription domain.Subscription
	err := s.db.GetContext(ctx, &subscription, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Subscription{}, domain.ErrNotFound
		}
		return domain.Subscription{}, err
	}

	return subscription, nil
}

func (s *SubscriptionRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
