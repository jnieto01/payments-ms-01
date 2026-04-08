package entity

import "time"

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

type Subscription struct {
	ID         uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	ClubID     string             `gorm:"type:varchar(100);not null;uniqueIndex" json:"club_id"`
	Plan       string             `gorm:"type:varchar(50);not null" json:"plan"`
	Status     SubscriptionStatus `gorm:"type:varchar(50);not null;default:active" json:"status"`
	PaymentID  *uint              `gorm:"index"                    json:"payment_id,omitempty"`
	Payment    *Payment           `gorm:"foreignKey:PaymentID"     json:"payment,omitempty"`
	StartsAt   time.Time          `gorm:"not null"                 json:"starts_at"`
	EndsAt     time.Time          `gorm:"not null"                 json:"ends_at"`
	CreatedAt  time.Time          `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt  time.Time          `gorm:"autoUpdateTime"           json:"updated_at"`
}
