package repository

import (
	"context"
	"errors"

	"github.com/jnieto01/payments-ms/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms/internal/domain/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type subscriptionRepositoryImpl struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) domainrepo.SubscriptionRepository {
	return &subscriptionRepositoryImpl{db: db}
}

func (r *subscriptionRepositoryImpl) Upsert(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "club_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"plan", "status", "payment_id", "starts_at", "ends_at", "updated_at"}),
		}).
		Create(sub).Error
}

func (r *subscriptionRepositoryImpl) GetByClubID(ctx context.Context, clubID string) (*entity.Subscription, error) {
	var s entity.Subscription
	err := r.db.WithContext(ctx).Where("club_id = ?", clubID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}
