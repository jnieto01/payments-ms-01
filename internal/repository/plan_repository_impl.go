package repository

import (
	"context"
	"errors"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms-01/internal/domain/repository"
	"gorm.io/gorm"
)

type planRepositoryImpl struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) domainrepo.PlanRepository {
	return &planRepositoryImpl{db: db}
}

func (r *planRepositoryImpl) GetAll(ctx context.Context) ([]entity.Plan, error) {
	var plans []entity.Plan
	err := r.db.WithContext(ctx).Where("is_active = 1").Order("price ASC").Find(&plans).Error
	return plans, err
}

func (r *planRepositoryImpl) GetByKey(ctx context.Context, key string) (*entity.Plan, error) {
	var plan entity.Plan
	err := r.db.WithContext(ctx).Where("plan_key = ? AND is_active = 1", key).First(&plan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &plan, err
}

func (r *planRepositoryImpl) UpdateMPPreapprovalPlan(ctx context.Context, planID uint, mpPlanID, mpPlanURL string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", planID).
		Updates(map[string]any{
			"mp_preapproval_plan_id":  mpPlanID,
			"mp_preapproval_plan_url": mpPlanURL,
		}).Error
}
