package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms-01/internal/domain/repository"
	"gorm.io/gorm"
)

type paymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) domainrepo.PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

func (r *paymentRepositoryImpl) Create(ctx context.Context, payment *entity.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepositoryImpl) GetByID(ctx context.Context, id uint) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.WithContext(ctx).First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *paymentRepositoryImpl) GetByIdempotencyKey(ctx context.Context, key string) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *paymentRepositoryImpl) GetByMPPaymentID(ctx context.Context, mpPaymentID int64) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.WithContext(ctx).Where("mp_payment_id = ?", mpPaymentID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *paymentRepositoryImpl) GetByExternalRef(ctx context.Context, externalRef string) (*entity.Payment, error) {
	var p entity.Payment
	err := r.db.WithContext(ctx).Where("external_ref = ?", externalRef).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *paymentRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status entity.PaymentStatus, mpPaymentID *int64) error {
	updates := map[string]interface{}{"status": status}
	if mpPaymentID != nil {
		updates["mp_payment_id"] = *mpPaymentID
	}
	return r.db.WithContext(ctx).Model(&entity.Payment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *paymentRepositoryImpl) ListByDateRange(ctx context.Context, from, to time.Time) ([]entity.Payment, error) {
	var payments []entity.Payment
	err := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at <= ?", from, to).
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}
