package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/programmerpark/subscription-aggregator/internal/model"
	"github.com/programmerpark/subscription-aggregator/internal/repository"
)

var (
	ErrNotFound     = repository.ErrNotFound
	ErrInvalidInput = errors.New("invalid input")
)

type subscriptionService struct {
	repo repository.SubscriptionRepository
	log  *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, log *slog.Logger) SubscriptionService {
	return &subscriptionService{repo: repo, log: log}
}

func (s *subscriptionService) Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	sub, err := s.repo.Create(ctx, req)
	if err != nil {
		s.log.Error("failed to create subscription", "error", err)
		return nil, err
	}
	s.log.Info("subscription created", "id", sub.ID, "user_id", sub.UserID, "service_name", sub.ServiceName)
	return sub, nil
}

func (s *subscriptionService) Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Error("failed to get subscription", "id", id, "error", err)
		}
		return nil, err
	}
	return sub, nil
}

func (s *subscriptionService) List(ctx context.Context, f model.ListFilter) (*model.ListResponse, error) {
	if f.UserID == uuid.Nil {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		s.log.Error("failed to list subscriptions", "error", err, "user_id", f.UserID)
		return nil, err
	}
	s.log.Debug("subscriptions listed", "user_id", f.UserID, "count", len(items), "total", total)
	return &model.ListResponse{
		Items:  items,
		Total:  total,
		Limit:  f.Limit,
		Offset: f.Offset,
	}, nil
}

func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.Subscription, error) {
	if err := validateUpdate(req); err != nil {
		return nil, err
	}

	sub, err := s.repo.Update(ctx, id, req)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Error("failed to update subscription", "id", id, "error", err)
		}
		return nil, err
	}
	s.log.Info("subscription updated", "id", sub.ID)
	return sub, nil
}

func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			s.log.Error("failed to delete subscription", "id", id, "error", err)
		}
		return err
	}
	s.log.Info("subscription deleted", "id", id)
	return nil
}

func (s *subscriptionService) Sum(ctx context.Context, f model.SumFilter) (*model.SumResponse, error) {
	if f.From.Time.After(f.To.Time) {
		return nil, fmt.Errorf("%w: from must be before or equal to to", ErrInvalidInput)
	}

	items, err := s.repo.ListForSum(ctx, f)
	if err != nil {
		s.log.Error("failed to calculate sum", "error", err)
		return nil, err
	}

	total := 0
	for _, item := range items {
		months := model.OverlapMonths(item.StartDate, item.EndDate, f.From, f.To)
		total += item.Price * months
	}

	s.log.Info("subscription sum calculated",
		"total", total,
		"from", f.From.String(),
		"to", f.To.String(),
		"matched", len(items),
	)

	return &model.SumResponse{
		Total:       total,
		From:        f.From,
		To:          f.To,
		UserID:      f.UserID,
		ServiceName: f.ServiceName,
	}, nil
}

func validateCreate(req model.CreateSubscriptionRequest) error {
	if req.ServiceName == "" {
		return fmt.Errorf("%w: service_name is required", ErrInvalidInput)
	}
	if req.Price < 0 {
		return fmt.Errorf("%w: price must be >= 0", ErrInvalidInput)
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if req.StartDate.IsZero() {
		return fmt.Errorf("%w: start_date is required (MM-YYYY)", ErrInvalidInput)
	}
	if req.EndDate != nil && req.EndDate.Time.Before(req.StartDate.Time) {
		return fmt.Errorf("%w: end_date must be >= start_date", ErrInvalidInput)
	}
	return nil
}

func validateUpdate(req model.UpdateSubscriptionRequest) error {
	if req.ServiceName != nil && *req.ServiceName == "" {
		return fmt.Errorf("%w: service_name cannot be empty", ErrInvalidInput)
	}
	if req.Price != nil && *req.Price < 0 {
		return fmt.Errorf("%w: price must be >= 0", ErrInvalidInput)
	}
	if req.UserID != nil && *req.UserID == uuid.Nil {
		return fmt.Errorf("%w: user_id cannot be empty", ErrInvalidInput)
	}
	if req.StartDate != nil && req.EndDate != nil && req.EndDate.Time.Before(req.StartDate.Time) {
		return fmt.Errorf("%w: end_date must be >= start_date", ErrInvalidInput)
	}
	return nil
}
