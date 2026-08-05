package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/programmerpark/subscription-aggregator/internal/model"
)

// SubscriptionService defines business operations for subscriptions.
type SubscriptionService interface {
	Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error)
	Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, f model.ListFilter) (*model.ListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Sum(ctx context.Context, f model.SumFilter) (*model.SumResponse, error)
}
