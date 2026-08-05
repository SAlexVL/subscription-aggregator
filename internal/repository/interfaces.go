package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/programmerpark/subscription-aggregator/internal/model"
)

// SubscriptionRepository defines persistence operations for subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, f model.ListFilter) ([]model.Subscription, int, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Sum(ctx context.Context, f model.SumFilter) (int, error)
}
