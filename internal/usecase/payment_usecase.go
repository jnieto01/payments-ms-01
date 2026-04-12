package usecase

import (
	"context"

	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
)

type CheckoutRequest struct {
	ClubID         string  `json:"club_id"         validate:"required"`
	UserID         string  `json:"user_id"         validate:"required"`
	PayerEmail     string  `json:"payer_email,omitempty"` // populated from JWT context by the handler
	Plan           string  `json:"plan"            validate:"required,oneof=pro avanzado"`
	IdempotencyKey string  `json:"idempotency_key" validate:"required"`
	Amount         float64 `json:"amount"`
}

// CheckoutType indicates whether the checkout is a one-time preference or a recurring preapproval.
type CheckoutType string

const (
	CheckoutTypePreference  CheckoutType = "preference"
	CheckoutTypePreapproval CheckoutType = "preapproval"
)

type CheckoutResponse struct {
	PaymentID       uint         `json:"payment_id"`
	InitPoint       string       `json:"init_point"`
	CheckoutType    CheckoutType `json:"checkout_type"`
	// PreferenceID is set when CheckoutType == "preference" (one-time payment fallback).
	PreferenceID    string       `json:"preference_id,omitempty"`
	// MPPreapprovalID is set when CheckoutType == "preapproval" (recurring subscription).
	MPPreapprovalID string       `json:"mp_preapproval_id,omitempty"`
}

type TrialRequest struct {
	ClubID string `json:"club_id" validate:"required"`
	Plan   string `json:"plan"    validate:"required,oneof=pro avanzado"`
}

// TrialStatusResponse is returned by the subscription-status endpoint.
type TrialStatusResponse struct {
	Plan                 string  `json:"plan"`                    // "basico" | "pro" | "avanzado"
	Status               string  `json:"status"`                  // "none" | "trial" | "active" | "expired" | "canceled"
	IsTrial              bool    `json:"is_trial"`
	DaysRemaining        int     `json:"days_remaining"`          // -1 if not applicable
	ShowWarning          bool    `json:"show_warning"`
	WarningThresholdDays int     `json:"warning_threshold_days"`
	TrialEndsAt          *string `json:"trial_ends_at,omitempty"`
	EndsAt               *string `json:"ends_at,omitempty"`
}

// WebhookPayload covers both payment and subscription events from MercadoPago.
// MP sends: {"action":"payment.updated","data":{"id":"12345"}}
// Also: {"action":"subscription_preapproval.updated","data":{"id":"..."}}
//        {"action":"subscription_authorized_payment.updated","data":{"id":"..."}}
type WebhookPayload struct {
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
	// MP also sends these top-level fields on some events
	Type string `json:"type"`
}

// MarketplaceCommissionResponse is the public shape returned to the frontend.
type MarketplaceCommissionResponse struct {
	Percentage float64 `json:"percentage"`  // e.g. 0.02
	WordingES  string  `json:"wording_es"`
	WordingEN  string  `json:"wording_en"`
	Currency   string  `json:"currency"`
}

// MarketplaceCheckoutRequest is sent by the frontend when the user publishes an item.
type MarketplaceCheckoutRequest struct {
	UserID         string  `json:"user_id"`
	PayerEmail     string  `json:"payer_email,omitempty"` // enriched from JWT by the handler
	ItemID         int64   `json:"item_id"    validate:"required,gt=0"`
	ItemTitle      string  `json:"item_title"  validate:"required"`
	Price          float64 `json:"price"       validate:"required,gt=0"`
	IdempotencyKey string  `json:"idempotency_key" validate:"required"`
}

// MarketplaceCheckoutResponse is returned after creating the MP preference.
type MarketplaceCheckoutResponse struct {
	PaymentID      uint    `json:"payment_id"`
	CommissionAmt  float64 `json:"commission_amount"`
	InitPoint      string  `json:"init_point"`
	PreferenceID   string  `json:"preference_id"`
}

type PaymentUsecase interface {
	CreateSubscriptionCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, payload WebhookPayload, xSignature, xRequestID string) error
	GetTrialStatus(ctx context.Context, clubID string) (*TrialStatusResponse, error)
	StartTrial(ctx context.Context, req TrialRequest) (*entity.Trial, error)
	GetPlans(ctx context.Context) ([]entity.Plan, error)
	GetMarketplaceCommission(ctx context.Context) (*MarketplaceCommissionResponse, error)
	CreateMarketplaceCheckout(ctx context.Context, req MarketplaceCheckoutRequest) (*MarketplaceCheckoutResponse, error)
}
