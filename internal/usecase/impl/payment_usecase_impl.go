package impl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jnieto01/payments-ms/internal/config"
	"github.com/jnieto01/payments-ms/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms/internal/domain/repository"
	"github.com/jnieto01/payments-ms/internal/usecase"
	"github.com/redis/go-redis/v9"
)

var planPrices = map[string]float64{
	"pro":      4999,
	"avanzado": 9999,
}

var planLabels = map[string]string{
	"pro":      "NVF Plan Pro — Licencia mensual",
	"avanzado": "NVF Plan Avanzado — Licencia mensual",
}

type paymentUsecaseImpl struct {
	paymentRepo      domainrepo.PaymentRepository
	subscriptionRepo domainrepo.SubscriptionRepository
	redis            *redis.Client
	mp               *mpClient
	mpCfg            config.MercadoPagoConfig
}

func NewPaymentUsecase(
	paymentRepo domainrepo.PaymentRepository,
	subscriptionRepo domainrepo.SubscriptionRepository,
	redisClient *redis.Client,
	mpCfg config.MercadoPagoConfig,
) usecase.PaymentUsecase {
	return &paymentUsecaseImpl{
		paymentRepo:      paymentRepo,
		subscriptionRepo: subscriptionRepo,
		redis:            redisClient,
		mp:               newMPClient(mpCfg),
		mpCfg:            mpCfg,
	}
}

func (u *paymentUsecaseImpl) CreateSubscriptionCheckout(ctx context.Context, req usecase.CheckoutRequest) (*usecase.CheckoutResponse, error) {
	// 1. DB idempotency check — fast path before acquiring lock
	existing, err := u.paymentRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if existing != nil {
		return &usecase.CheckoutResponse{
			PaymentID:    existing.ID,
			InitPoint:    existing.InitPoint,
			PreferenceID: existing.MPPreferenceID,
		}, nil
	}

	// 2. Distributed lock via Redis SetNX (prevents concurrent duplicate requests)
	lockKey := fmt.Sprintf("checkout_lock:%s", req.IdempotencyKey)
	locked, err := u.redis.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lock: %w", err)
	}
	if !locked {
		// Another instance is processing — wait briefly and double-check DB
		time.Sleep(200 * time.Millisecond)
		existing, err = u.paymentRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
		if err != nil {
			return nil, fmt.Errorf("idempotency recheck: %w", err)
		}
		if existing != nil {
			return &usecase.CheckoutResponse{
				PaymentID:    existing.ID,
				InitPoint:    existing.InitPoint,
				PreferenceID: existing.MPPreferenceID,
			}, nil
		}
		return nil, fmt.Errorf("payment is already being processed, please retry in a moment")
	}
	defer u.redis.Del(ctx, lockKey)

	// 3. Double-check DB after acquiring lock (handles race between lock check and DB read)
	existing, err = u.paymentRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency double-check: %w", err)
	}
	if existing != nil {
		return &usecase.CheckoutResponse{
			PaymentID:    existing.ID,
			InitPoint:    existing.InitPoint,
			PreferenceID: existing.MPPreferenceID,
		}, nil
	}

	// 4. Determine amount
	amount := planPrices[req.Plan]
	if amount == 0 {
		return nil, fmt.Errorf("invalid plan: %s", req.Plan)
	}
	label := planLabels[req.Plan]

	// 5. Create preference in MercadoPago
	externalRef := fmt.Sprintf("club_%s_plan_%s_%s", req.ClubID, req.Plan, req.IdempotencyKey)
	prefReq := MPPreferenceRequest{
		Items: []MPPreferenceItem{{
			Title:      label,
			Quantity:   1,
			UnitPrice:  amount,
			CurrencyID: "ARS",
		}},
		ExternalReference: externalRef,
		NotificationURL:   u.mpCfg.NotificationURL,
	}

	pref, err := u.mp.CreatePreference(ctx, prefReq)
	if err != nil {
		return nil, fmt.Errorf("create MP preference: %w", err)
	}

	// 6. Persist payment record
	payment := &entity.Payment{
		ClubID:         req.ClubID,
		UserID:         req.UserID,
		Type:           entity.PaymentTypeSubscription,
		Plan:           req.Plan,
		Amount:         amount,
		Currency:       "ARS",
		Status:         entity.PaymentStatusPending,
		IdempotencyKey: req.IdempotencyKey,
		MPPreferenceID: pref.ID,
		InitPoint:      pref.InitPoint,
		ExternalRef:    externalRef,
	}
	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	return &usecase.CheckoutResponse{
		PaymentID:    payment.ID,
		InitPoint:    pref.InitPoint,
		PreferenceID: pref.ID,
	}, nil
}

func (u *paymentUsecaseImpl) HandleWebhook(ctx context.Context, payload usecase.WebhookPayload, signature string) error {
	if payload.Action != "payment.updated" && payload.Action != "payment.created" {
		return nil
	}

	mpPaymentIDStr := payload.Data.ID
	if mpPaymentIDStr == "" {
		return nil
	}

	// Webhook deduplication — prevent reprocessing same notification
	dedupKey := fmt.Sprintf("webhook:%s:%s", mpPaymentIDStr, payload.Action)
	set, err := u.redis.SetNX(ctx, dedupKey, "1", 5*time.Minute).Result()
	if err == nil && !set {
		return nil // already processed
	}

	// Fetch payment details from MP
	mpResp, err := u.mp.GetPayment(ctx, mpPaymentIDStr)
	if err != nil {
		return fmt.Errorf("fetch MP payment: %w", err)
	}

	mpPaymentID, _ := strconv.ParseInt(mpPaymentIDStr, 10, 64)

	// Find our internal payment by external reference
	payment, err := u.paymentRepo.GetByMPPaymentID(ctx, mpPaymentID)
	if err != nil {
		return fmt.Errorf("find payment: %w", err)
	}

	// Try by external ref if not found by MP payment ID
	if payment == nil {
		return nil // payment not tracked (may be from another system)
	}

	// Map MP status to internal status
	var newStatus entity.PaymentStatus
	switch mpResp.Status {
	case "approved":
		newStatus = entity.PaymentStatusApproved
	case "rejected":
		newStatus = entity.PaymentStatusRejected
	case "cancelled":
		newStatus = entity.PaymentStatusCanceled
	default:
		newStatus = entity.PaymentStatusPending
	}

	if err := u.paymentRepo.UpdateStatus(ctx, payment.ID, newStatus, &mpPaymentID); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	// If approved, activate subscription
	if newStatus == entity.PaymentStatusApproved && payment.Type == entity.PaymentTypeSubscription {
		now := time.Now()
		sub := &entity.Subscription{
			ClubID:    payment.ClubID,
			Plan:      payment.Plan,
			Status:    entity.SubscriptionStatusActive,
			PaymentID: &payment.ID,
			StartsAt:  now,
			EndsAt:    now.AddDate(0, 1, 0), // 1 month
		}
		if err := u.subscriptionRepo.Upsert(ctx, sub); err != nil {
			return fmt.Errorf("upsert subscription: %w", err)
		}
	}

	return nil
}

func (u *paymentUsecaseImpl) GetSubscriptionByClub(ctx context.Context, clubID string) (*entity.Subscription, error) {
	return u.subscriptionRepo.GetByClubID(ctx, clubID)
}
