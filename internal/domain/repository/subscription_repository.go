package repository

import (
	"context"

	"github.com/jnieto01/payments-ms/internal/domain/entity"
)

type SubscriptionRepository interface {
	Upsert(ctx context.Context, sub *entity.Subscription) error
	GetByClubID(ctx context.Context, clubID string) (*entity.Subscription, error)
}
