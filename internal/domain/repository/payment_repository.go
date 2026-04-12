package repository

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*entity.Payment, error)
	GetByMPPaymentID(ctx context.Context, mpPaymentID int64) (*entity.Payment, error)
	GetByExternalRef(ctx context.Context, externalRef string) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, id uint, status entity.PaymentStatus, mpPaymentID *int64) error
}
