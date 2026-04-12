package repository

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type PlanRepository interface {
	GetAll(ctx context.Context) ([]entity.Plan, error)
	GetByKey(ctx context.Context, key string) (*entity.Plan, error)
	UpdateMPPreapprovalPlan(ctx context.Context, planID uint, mpPlanID, mpPlanURL string) error
}
