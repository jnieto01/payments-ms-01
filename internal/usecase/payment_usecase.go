package usecase

import (
	"context"

	"github.com/jnieto01/payments-ms/internal/domain/entity"
)

type CheckoutRequest struct {
	ClubID         string  `json:"club_id"          validate:"required"`
	UserID         string  `json:"user_id"          validate:"required"`
	Plan           string  `json:"plan"             validate:"required,oneof=pro avanzado"`
	IdempotencyKey string  `json:"idempotency_key"  validate:"required,uuid4"`
	Amount         float64 `json:"amount"`
}

type CheckoutResponse struct {
	PaymentID  uint   `json:"payment_id"`
	InitPoint  string `json:"init_point"`
	PreferenceID string `json:"preference_id"`
}

type WebhookPayload struct {
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

type PaymentUsecase interface {
	CreateSubscriptionCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, payload WebhookPayload, signature string) error
	GetSubscriptionByClub(ctx context.Context, clubID string) (*entity.Subscription, error)
}
