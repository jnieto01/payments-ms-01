package repository

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type PlanRepository interface {
	GetAll(ctx context.Context) ([]entity.Plan, error)
	GetAllAdmin(ctx context.Context) ([]entity.Plan, error)
	GetByKey(ctx context.Context, key string) (*entity.Plan, error)
	GetByID(ctx context.Context, id uint) (*entity.Plan, error)
	Create(ctx context.Context, plan *entity.Plan) error
	Update(ctx context.Context, id uint, fields map[string]any) error
	SoftDelete(ctx context.Context, id uint) error
	UpdateMPPreapprovalPlan(ctx context.Context, planID uint, mpPlanID, mpPlanURL string) error
}
