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

func (r *planRepositoryImpl) GetAllAdmin(ctx context.Context) ([]entity.Plan, error) {
	var plans []entity.Plan
	err := r.db.WithContext(ctx).Order("price ASC").Find(&plans).Error
	return plans, err
}

func (r *planRepositoryImpl) GetByID(ctx context.Context, id uint) (*entity.Plan, error) {
	var plan entity.Plan
	err := r.db.WithContext(ctx).First(&plan, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &plan, err
}

func (r *planRepositoryImpl) Create(ctx context.Context, plan *entity.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepositoryImpl) Update(ctx context.Context, id uint, fields map[string]any) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", id).
		Updates(fields).Error
}

func (r *planRepositoryImpl) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&entity.Plan{}).
		Where("id = ?", id).
		Update("is_active", false).Error
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
