package repository

import (
	"context"
	"time"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*entity.Payment, error)
	GetByMPPaymentID(ctx context.Context, mpPaymentID int64) (*entity.Payment, error)
	GetByExternalRef(ctx context.Context, externalRef string) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, id uint, status entity.PaymentStatus, mpPaymentID *int64) error
	ListByDateRange(ctx context.Context, from, to time.Time) ([]entity.Payment, error)
	// GetApprovedByItemID returns the first approved payment for a given market item, or nil.
	GetApprovedByItemID(ctx context.Context, itemID int64) (*entity.Payment, error)
}
