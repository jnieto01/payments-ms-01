package entity

import "time"

type TrialStatus string

const (
	TrialStatusActive   TrialStatus = "active"
	TrialStatusTrial    TrialStatus = "trial"
	TrialStatusCanceled TrialStatus = "canceled"
	TrialStatusExpired  TrialStatus = "expired"
)

type Trial struct {
	ID                  uint        `gorm:"primaryKey;autoIncrement"              json:"id"`
	ClubID              string      `gorm:"type:varchar(100);not null;uniqueIndex" json:"club_id"`
	Plan                string      `gorm:"type:varchar(50);not null"              json:"plan"`
	Status              TrialStatus `gorm:"type:varchar(50);not null;default:active" json:"status"`
	IsTrial             bool        `gorm:"not null;default:false"                json:"is_trial"`
	TrialStartsAt       *time.Time  `json:"trial_starts_at,omitempty"`
	TrialEndsAt         *time.Time  `json:"trial_ends_at,omitempty"`
	PaymentID           *uint       `gorm:"index"                                 json:"payment_id,omitempty"`
	Payment             *Payment    `gorm:"foreignKey:PaymentID"                  json:"payment,omitempty"`
	MPPreapprovalID     string      `gorm:"type:varchar(255)"                     json:"mp_preapproval_id,omitempty"`
	MPPreapprovalStatus string      `gorm:"type:varchar(50)"                      json:"mp_preapproval_status,omitempty"`
	StartsAt            time.Time   `gorm:"not null"                              json:"starts_at"`
	EndsAt              time.Time   `gorm:"not null"                              json:"ends_at"`
	CreatedAt           time.Time   `gorm:"autoCreateTime"                        json:"created_at"`
	UpdatedAt           time.Time   `gorm:"autoUpdateTime"                        json:"updated_at"`
}

// TableName tells GORM to use the `trials` table.
func (Trial) TableName() string { return "trials" }
