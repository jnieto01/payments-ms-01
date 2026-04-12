package repository

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type TrialRepository interface {
	Upsert(ctx context.Context, trial *entity.Trial) error
	GetByClubID(ctx context.Context, clubID string) (*entity.Trial, error)
	UpdateStatus(ctx context.Context, id uint, status entity.TrialStatus) error
	UpdatePreapproval(ctx context.Context, id uint, mpPreapprovalID, mpStatus string) error
}
