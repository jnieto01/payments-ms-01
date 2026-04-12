package repository

import (
	"context"
	"errors"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms-01/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type trialRepositoryImpl struct {
	db *gorm.DB
}

func NewTrialRepository(db *gorm.DB) domainrepo.TrialRepository {
	return &trialRepositoryImpl{db: db}
}

func (r *trialRepositoryImpl) Upsert(ctx context.Context, trial *entity.Trial) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "club_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"plan", "status", "is_trial", "trial_starts_at", "trial_ends_at",
				"payment_id", "mp_preapproval_id", "mp_preapproval_status",
				"starts_at", "ends_at", "updated_at",
			}),
		}).
		Create(trial).Error
}

func (r *trialRepositoryImpl) GetByClubID(ctx context.Context, clubID string) (*entity.Trial, error) {
	var t entity.Trial
	err := r.db.WithContext(ctx).Where("club_id = ?", clubID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *trialRepositoryImpl) UpdateStatus(ctx context.Context, id uint, status entity.TrialStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.Trial{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *trialRepositoryImpl) UpdatePreapproval(ctx context.Context, id uint, mpPreapprovalID, mpStatus string) error {
	return r.db.WithContext(ctx).
		Model(&entity.Trial{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"mp_preapproval_id":     mpPreapprovalID,
			"mp_preapproval_status": mpStatus,
		}).Error
}
