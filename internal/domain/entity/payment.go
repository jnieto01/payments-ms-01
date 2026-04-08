package entity

import "time"

type PaymentType string
type PaymentStatus string

const (
	PaymentTypeSubscription PaymentType = "subscription"
	PaymentTypeAdvertising  PaymentType = "advertising"

	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusApproved PaymentStatus = "approved"
	PaymentStatusRejected PaymentStatus = "rejected"
	PaymentStatusCanceled PaymentStatus = "canceled"
)

type Payment struct {
	ID             uint          `gorm:"primaryKey;autoIncrement"           json:"id"`
	ClubID         string        `gorm:"type:varchar(100);not null;index"   json:"club_id"`
	UserID         string        `gorm:"type:varchar(100);not null;index"   json:"user_id"`
	Type           PaymentType   `gorm:"type:varchar(50);not null"          json:"type"`
	Plan           string        `gorm:"type:varchar(50)"                   json:"plan,omitempty"`
	Amount         float64       `gorm:"type:decimal(12,2);not null"        json:"amount"`
	Currency       string        `gorm:"type:varchar(10);not null;default:ARS" json:"currency"`
	Status         PaymentStatus `gorm:"type:varchar(50);not null;default:pending" json:"status"`
	IdempotencyKey string        `gorm:"type:varchar(100);not null;uniqueIndex" json:"idempotency_key"`
	MPPaymentID    *int64        `gorm:"index"                              json:"mp_payment_id,omitempty"`
	MPPreferenceID string        `gorm:"type:varchar(255)"                  json:"mp_preference_id,omitempty"`
	InitPoint      string        `gorm:"type:text"                          json:"init_point,omitempty"`
	ExternalRef    string        `gorm:"type:varchar(255);index"            json:"external_ref,omitempty"`
	CreatedAt      time.Time     `gorm:"autoCreateTime"                     json:"created_at"`
	UpdatedAt      time.Time     `gorm:"autoUpdateTime"                     json:"updated_at"`
}
