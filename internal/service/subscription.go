package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	contextkeys "tz/internal/contextkey"
	"tz/internal/domain"
	"tz/internal/dto"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionRepositoryI interface {
	CreateSubscription(ctx context.Context, sub domain.Subscription) (domain.Subscription, error)
	SubscriptionByID(ctx context.Context, id uuid.UUID) (domain.Subscription, error)
	Subscriptions(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, int, error)
	SubscriptionsCost(ctx context.Context, filter domain.CostRequest) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, sub domain.UpdateSubscription) (domain.Subscription, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
}

type SubscriptionService struct {
	repo SubscriptionRepositoryI
	log  *zap.Logger
}

func NewSubscriptionService(repo SubscriptionRepositoryI, log *zap.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, req dto.CreateSubscriptionRequest) (dto.SubscriptionOutput, error) {
	log := s.loggerWith(ctx, zap.String("user_id", req.UserID), zap.String("service_name", req.ServiceName))

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		log.Warn("Invalid start_date format", zap.String("date", req.StartDate))
		return dto.SubscriptionOutput{}, fmt.Errorf("%w: %s", err, req.StartDate)
	}

	var endDate *time.Time
	if req.EndDate != nil {
		ed, err := parseDate(*req.EndDate)
		if err != nil {
			log.Warn("Invalid end_date format", zap.String("date", *req.EndDate))
			return dto.SubscriptionOutput{}, fmt.Errorf("%w: %s", err, *req.EndDate)
		}
		if ed.Before(startDate) {
			log.Warn("End date before start date")
			return dto.SubscriptionOutput{}, fmt.Errorf("end_date must be after start_date. End  date: %s, Start date: %s", ed, startDate)
		}
		endDate = &ed
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Warn("Invalid user_id format", zap.String("user_id", req.UserID))
		return dto.SubscriptionOutput{}, fmt.Errorf("invalid user_id: %w", err)
	}

	id := uuid.New()
	subscription := domain.Subscription{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      userID,
	}

	subscriptionDB, err := s.repo.CreateSubscription(ctx, subscription)
	if err != nil {
		log.Error("Failed to create subscription", zap.Error(err))
		return dto.SubscriptionOutput{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	log.Info("Subscription created", zap.String("subscription_id", subscriptionDB.ID.String()), zap.String("created_at", subscriptionDB.CreatedAt.Format(time.DateOnly)))
	return subscriptionToDto(subscriptionDB), nil
}

func (s *SubscriptionService) SubscriptionByID(ctx context.Context, id uuid.UUID) (dto.SubscriptionOutput, error) {
	log := s.loggerWith(ctx, zap.String("subscription_id", id.String()))

	subscriptionDB, err := s.repo.SubscriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Subscription not found")
			return dto.SubscriptionOutput{}, domain.ErrNotFound
		}
		log.Error("Failed to get subscription", zap.Error(err))
		return dto.SubscriptionOutput{}, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscriptionToDto(subscriptionDB), nil
}

func (s *SubscriptionService) Subscriptions(ctx context.Context, filter dto.SubscriptionFilter) (dto.SubscriptionsOutput, error) {
	log := s.loggerWith(ctx)
	var userID *uuid.UUID
	if filter.UserID != nil {
		uid, err := uuid.Parse(*filter.UserID)
		if err != nil {
			log.Warn("Invalid user_id in filter", zap.String("user_id", *filter.UserID), zap.Error(err))
			return dto.SubscriptionsOutput{}, fmt.Errorf("invalid user_id in filter: %w", err)
		}
		userID = &uid
	}

	subscriptionsDB, total, err := s.repo.Subscriptions(ctx, domain.SubscriptionFilter{
		UserID:      userID,
		ServiceName: filter.ServiceName,
		Limit:       filter.PageSize,
		Offset:      (filter.Page - 1) * filter.PageSize,
	})

	if err != nil {
		log.Error("Failed to fetch subscriptions", zap.Error(err))
		return dto.SubscriptionsOutput{}, fmt.Errorf("failed to fetch subscriptions: %w", err)
	}

	subscriptions := make([]dto.SubscriptionOutput, len(subscriptionsDB))
	for i, subscription := range subscriptionsDB {
		subscriptions[i] = subscriptionToDto(subscription)
	}

	log.Info("Subscriptions fetched", zap.Int("count", len(subscriptions)), zap.Int("total", total))
	return dto.MakeSubscriptionsOutput(subscriptions, total, filter.Page, filter.PageSize), nil
}

func (s *SubscriptionService) SubscriptionsCost(ctx context.Context, filter dto.CostRequest) (int, error) {
	log := s.loggerWith(ctx)

	var userID *uuid.UUID
	if filter.UserID != nil {
		uid, err := uuid.Parse(*filter.UserID)
		if err != nil {
			log.Warn("Invalid user_id in filter", zap.String("user_id", *filter.UserID), zap.Error(err))
			return 0, fmt.Errorf("invalid user_id in filter: %w", err)
		}
		userID = &uid
	}

	var startDate, endDate *time.Time
	if filter.From != nil {
		sd, err := parseDate(*filter.From)
		if err != nil {
			log.Warn("Invalid start_date in update", zap.Error(err))
			return 0, fmt.Errorf("invalid start_date: %w", err)
		}
		startDate = &sd
	}
	if filter.To != nil {
		ed, err := parseDate(*filter.To)
		if err != nil {
			log.Warn("Invalid end_date in update", zap.Error(err))
			return 0, fmt.Errorf("invalid end_date: %w", err)
		}
		endDate = &ed
	}

	subs, err := s.repo.SubscriptionsCost(ctx, domain.CostRequest{
		ServiceName: filter.ServiceName,
		UserID:      userID,
	})
	if err != nil {
		log.Error("Error getting subscriptions cost", zap.Error(err))
		return 0, fmt.Errorf("error getting subscriptions cost: %w", err)
	}

	total := 0
	now := time.Now()
	for _, sub := range subs {
		// Определяем реальный период активности подписки
		actualStart := sub.StartDate
		actualEnd := now
		if sub.EndDate != nil {
			actualEnd = *sub.EndDate
		}

		calcStart := actualStart
		if startDate != nil {
			calcStart = maxTime(actualStart, *startDate)
		}

		calcEnd := actualEnd
		if endDate != nil {
			calcEnd = minTime(actualEnd, *endDate)
		}

		if calcEnd.Before(calcStart) {
			continue
		}

		months := monthsBetween(calcStart, calcEnd)
		total += sub.Price * months
	}

	return total, nil
}
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req dto.UpdateSubscriptionRequest) (dto.SubscriptionOutput, error) {
	log := s.loggerWith(ctx, zap.String("subscription_id", id.String()))

	var startDate, endDate *time.Time
	if req.StartDate != nil {
		sd, err := parseDate(*req.StartDate)
		if err != nil {
			log.Warn("Invalid start_date in update", zap.String("date", *req.StartDate))
			return dto.SubscriptionOutput{}, fmt.Errorf("%w: %s", err, *req.StartDate)
		}
		startDate = &sd
	}
	if req.EndDate != nil {
		ed, err := parseDate(*req.EndDate)
		if err != nil {
			log.Warn("Invalid end_date in update", zap.String("date", *req.EndDate))
			return dto.SubscriptionOutput{}, fmt.Errorf("%w: %s", err, *req.EndDate)
		}
		endDate = &ed
	}

	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		log.Warn("End date before start date in update")
		return dto.SubscriptionOutput{}, fmt.Errorf("end_date must be after start_date. End  date: %s, Start date: %s", *endDate, *startDate)
	}

	subscriptionDB, err := s.repo.UpdateSubscription(ctx, id, domain.UpdateSubscription{
		Price:       req.Price,
		ServiceName: req.ServiceName,
		StartDate:   startDate,
		EndDate:     endDate,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Subscription not found")
			return dto.SubscriptionOutput{}, fmt.Errorf("subscription not found: %w", domain.ErrNotFound)
		}
		log.Error("Failed to update subscription", zap.Error(err))
		return dto.SubscriptionOutput{}, fmt.Errorf("failed to update subscription: %w", err)
	}

	log.Info("Subscription updated")
	return subscriptionToDto(subscriptionDB), nil
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	log := s.loggerWith(ctx, zap.String("subscription_id", id.String()))

	if err := s.repo.DeleteSubscription(ctx, id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Warn("Subscription not found")
			return fmt.Errorf("subscription not found: %w", domain.ErrNotFound)
		}
		log.Error("Failed to delete subscription", zap.Error(err))
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	log.Info("Subscription deleted")
	return nil
}

func (s *SubscriptionService) loggerWith(ctx context.Context, fields ...zap.Field) *zap.Logger {
	requestID := ctx.Value(contextkeys.RequestIDKey)
	return s.log.With(append([]zap.Field{zap.String("request_id", requestID.(string))}, fields...)...)
}

func subscriptionToDto(s domain.Subscription) dto.SubscriptionOutput {
	var endDate *string
	if s.EndDate != nil {
		ed := parseTimeToString(*s.EndDate)
		endDate = &ed
	}
	return dto.SubscriptionOutput{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		StartDate:   parseTimeToString(s.StartDate),
		EndDate:     endDate,
		UserID:      s.UserID.String(),
		CreatedAt:   s.CreatedAt.Format(time.DateTime),
		UpdatedAt:   s.UpdatedAt.Format(time.DateTime),
	}
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func monthsBetween(start, end time.Time) int {
	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	if end.Before(start) {
		return 0
	}

	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month())
}

func parseDate(dateStr string) (time.Time, error) {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("%w: %v", domain.ErrInvalidDateFormat, dateStr)
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", domain.ErrInvalidDate, dateStr)
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", domain.ErrInvalidDate, dateStr)
	}
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("%w: %v", domain.ErrInvalidDate, dateStr)
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC), nil
}

func parseTimeToString(t time.Time) string {
	return fmt.Sprintf("%02d-%d", t.Month(), t.Year())
}
