package repository

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type MarketplaceCommissionRepository interface {
	GetActive(ctx context.Context) (*entity.MarketplaceCommission, error)
}
