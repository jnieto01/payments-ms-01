package repository

import (
	"context"
	"errors"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms-01/internal/domain/repository"
	"gorm.io/gorm"
)

type marketplaceCommissionRepositoryImpl struct {
	db *gorm.DB
}

func NewMarketplaceCommissionRepository(db *gorm.DB) domainrepo.MarketplaceCommissionRepository {
	return &marketplaceCommissionRepositoryImpl{db: db}
}

func (r *marketplaceCommissionRepositoryImpl) GetActive(ctx context.Context) (*entity.MarketplaceCommission, error) {
	var commission entity.MarketplaceCommission
	err := r.db.WithContext(ctx).
		Where("is_active = 1").
		Order("id ASC").
		First(&commission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &commission, err
}
