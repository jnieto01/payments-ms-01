package usecase

import (
	"context"
	"time"

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

// ManualPaymentRequest is used by the admin to manually approve a pending payment.
type ManualPaymentRequest struct {
	Notes string `json:"notes"`
}

// AdminAssignTrialRequest is the body for admin trial assignment (override-safe).
type AdminAssignTrialRequest struct {
	ClubID     string `json:"club_id"     validate:"required"`
	ClubNombre string `json:"club_nombre"`
	Plan       string `json:"plan"        validate:"required"`
	Days       int    `json:"dias"        validate:"required,gt=0"`
}

// AdminExtendTrialRequest is the body for trial extension.
type AdminExtendTrialRequest struct {
	Days int `json:"dias" validate:"required,gt=0"`
}

// AdminTrialResponse is the shape returned to the frontend for each trial row.
type AdminTrialResponse struct {
	ID            uint   `json:"id"`
	ClubID        string `json:"club_id"`
	ClubNombre    string `json:"club_nombre"`
	Plan          string `json:"plan"`
	FechaInicio   string `json:"fecha_inicio"`
	FechaFin      string `json:"fecha_fin"`
	DiasRestantes int    `json:"dias_restantes"`
	Estado        string `json:"estado"` // "activo" | "vencido" | "cancelado"
}

// CreatePlanRequest is the body for admin plan creation.
type CreatePlanRequest struct {
	PlanKey     string   `json:"plan_key"     validate:"required"`
	Name        string   `json:"name"         validate:"required"`
	ShortName   string   `json:"short_name"`
	Price       float64  `json:"price"        validate:"gte=0"`
	PriceUSD    float64  `json:"price_usd"`
	Currency    string   `json:"currency"`
	Period      string   `json:"period"`
	Features    []string `json:"features"`
	TrialDays   int      `json:"trial_days"`
	WarningDays int      `json:"warning_days"`
}

// UpdatePlanRequest allows partial updates — only non-nil fields are applied.
type UpdatePlanRequest struct {
	Name        *string  `json:"name,omitempty"`
	ShortName   *string  `json:"short_name,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	PriceUSD    *float64 `json:"price_usd,omitempty"`
	Currency    *string  `json:"currency,omitempty"`
	Period      *string  `json:"period,omitempty"`
	Features    []string `json:"features,omitempty"`
	TrialDays   *int     `json:"trial_days,omitempty"`
	WarningDays *int     `json:"warning_days,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// AdvertisingPaymentConfirmedMsg is the queue message published by sports-ms
// when an admin manually confirms an advertising payment (transfer or MP).
type AdvertisingPaymentConfirmedMsg struct {
	ItemID        int64   `json:"item_id"`
	SellerID      string  `json:"seller_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"` // transfer | mercado_pago
	ConfirmedAt   string  `json:"confirmed_at"`   // RFC3339
}

type PaymentUsecase interface {
	CreateSubscriptionCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	HandleWebhook(ctx context.Context, payload WebhookPayload, xSignature, xRequestID string) error
	GetTrialStatus(ctx context.Context, clubID string) (*TrialStatusResponse, error)
	StartTrial(ctx context.Context, req TrialRequest) (*entity.Trial, error)
	GetPlans(ctx context.Context) ([]entity.Plan, error)
	GetMarketplaceCommission(ctx context.Context) (*MarketplaceCommissionResponse, error)
	CreateMarketplaceCheckout(ctx context.Context, req MarketplaceCheckoutRequest) (*MarketplaceCheckoutResponse, error)
	// Admin-only payments
	ListPayments(ctx context.Context, from, to time.Time) ([]entity.Payment, error)
	RegisterManualPayment(ctx context.Context, paymentID uint, req ManualPaymentRequest) error
	HandleManualAdPaymentConfirmed(ctx context.Context, msg AdvertisingPaymentConfirmedMsg) error
	// Admin-only trial management
	AdminListTrials(ctx context.Context) ([]AdminTrialResponse, error)
	AdminAssignTrial(ctx context.Context, req AdminAssignTrialRequest) error
	AdminExtendTrial(ctx context.Context, id uint, days int) error
	AdminCancelTrial(ctx context.Context, id uint) error
	// Admin-only plan management
	AdminGetPlans(ctx context.Context) ([]entity.Plan, error)
	AdminGetPlanByID(ctx context.Context, id uint) (*entity.Plan, error)
	AdminCreatePlan(ctx context.Context, req CreatePlanRequest) (*entity.Plan, error)
	AdminUpdatePlan(ctx context.Context, id uint, req UpdatePlanRequest) (*entity.Plan, error)
	AdminTogglePlan(ctx context.Context, id uint) error
	AdminLinkMPPlan(ctx context.Context, id uint) error
}
