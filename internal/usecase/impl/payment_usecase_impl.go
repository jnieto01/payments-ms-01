package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	appconfig "github.com/jnieto01/payments-ms-01/internal/config"
	"github.com/jnieto01/payments-ms-01/internal/domain/entity"
	domainrepo "github.com/jnieto01/payments-ms-01/internal/domain/repository"
	"github.com/jnieto01/payments-ms-01/internal/usecase"
	"github.com/jnieto01/utils-01/rabbitmq"
	"github.com/redis/go-redis/v9"
)

// Compile-time check that all interface methods are implemented.
var _ usecase.PaymentUsecase = (*paymentUsecaseImpl)(nil)

const queueMarketplaceAdApproved = "marketplace.ad.approved"

type paymentUsecaseImpl struct {
	paymentRepo    domainrepo.PaymentRepository
	trialRepo      domainrepo.TrialRepository
	planRepo       domainrepo.PlanRepository
	commissionRepo domainrepo.MarketplaceCommissionRepository
	redis          *redis.Client
	mp             *mpClient
	mpCfg          appconfig.MercadoPagoConfig
	adPublisher    *rabbitmq.MessageService // nil when RabbitMQ not configured
}

func NewPaymentUsecase(
	paymentRepo domainrepo.PaymentRepository,
	trialRepo domainrepo.TrialRepository,
	planRepo domainrepo.PlanRepository,
	commissionRepo domainrepo.MarketplaceCommissionRepository,
	redisClient *redis.Client,
	mpCfg appconfig.MercadoPagoConfig,
	rmqCfg rabbitmq.Config,
) usecase.PaymentUsecase {
	var adPublisher *rabbitmq.MessageService
	if rmqCfg.Host != "" {
		pub, err := rabbitmq.NewMessageService(rmqCfg, queueMarketplaceAdApproved)
		if err != nil {
			fmt.Printf("WARN: could not connect to RabbitMQ for ad publisher: %v\n", err)
		} else {
			adPublisher = pub
		}
	}
	return &paymentUsecaseImpl{
		paymentRepo:    paymentRepo,
		trialRepo:      trialRepo,
		planRepo:       planRepo,
		commissionRepo: commissionRepo,
		redis:          redisClient,
		mp:             newMPClient(mpCfg),
		mpCfg:          mpCfg,
		adPublisher:    adPublisher,
	}
}

func (u *paymentUsecaseImpl) GetPlans(ctx context.Context) ([]entity.Plan, error) {
	return u.planRepo.GetAll(ctx)
}

func (u *paymentUsecaseImpl) GetTrialStatus(ctx context.Context, clubID string) (*usecase.TrialStatusResponse, error) {
	trial, err := u.trialRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// No trial at all → auto-activate avanzado trial
	if trial == nil {
		newTrial, err := u.StartTrial(ctx, usecase.TrialRequest{
			ClubID: clubID,
			Plan:   "avanzado",
		})
		if err != nil || newTrial == nil {
			return &usecase.TrialStatusResponse{
				Plan:                 "basico",
				Status:               "none",
				IsTrial:              false,
				DaysRemaining:        -1,
				ShowWarning:          false,
				WarningThresholdDays: 0,
			}, nil
		}
		trial = newTrial
	}

	// Fetch plan config for warning_days
	plan, _ := u.planRepo.GetByKey(ctx, trial.Plan)
	warningDays := 5
	if plan != nil {
		warningDays = plan.WarningDays
	}

	// Trial path
	if trial.IsTrial && trial.TrialEndsAt != nil {
		daysLeft := int(trial.TrialEndsAt.Sub(now).Hours() / 24)

		// Trial expired → revert to basico
		if now.After(*trial.TrialEndsAt) {
			trial.Status = entity.TrialStatusExpired
			_ = u.trialRepo.UpdateStatus(ctx, trial.ID, entity.TrialStatusExpired)
			return &usecase.TrialStatusResponse{
				Plan:                 "basico",
				Status:               "expired",
				IsTrial:              true,
				DaysRemaining:        0,
				ShowWarning:          false,
				WarningThresholdDays: warningDays,
			}, nil
		}

		trialEndsStr := trial.TrialEndsAt.Format(time.RFC3339)
		return &usecase.TrialStatusResponse{
			Plan:                 trial.Plan,
			Status:               "trial",
			IsTrial:              true,
			DaysRemaining:        daysLeft,
			ShowWarning:          daysLeft <= warningDays,
			WarningThresholdDays: warningDays,
			TrialEndsAt:          &trialEndsStr,
		}, nil
	}

	// Paid subscription (trial record promoted to active after payment)
	daysLeft := int(trial.EndsAt.Sub(now).Hours() / 24)

	// Expired or canceled → basico
	if trial.Status == entity.TrialStatusExpired || trial.Status == entity.TrialStatusCanceled {
		return &usecase.TrialStatusResponse{
			Plan:                 "basico",
			Status:               string(trial.Status),
			IsTrial:              false,
			DaysRemaining:        0,
			ShowWarning:          false,
			WarningThresholdDays: warningDays,
		}, nil
	}

	// Check if naturally expired
	if now.After(trial.EndsAt) {
		_ = u.trialRepo.UpdateStatus(ctx, trial.ID, entity.TrialStatusExpired)
		return &usecase.TrialStatusResponse{
			Plan:                 "basico",
			Status:               "expired",
			IsTrial:              false,
			DaysRemaining:        0,
			ShowWarning:          false,
			WarningThresholdDays: warningDays,
		}, nil
	}

	endsAtStr := trial.EndsAt.Format(time.RFC3339)
	return &usecase.TrialStatusResponse{
		Plan:                 trial.Plan,
		Status:               "active",
		IsTrial:              false,
		DaysRemaining:        daysLeft,
		ShowWarning:          daysLeft <= warningDays,
		WarningThresholdDays: warningDays,
		EndsAt:               &endsAtStr,
	}, nil
}

func (u *paymentUsecaseImpl) StartTrial(ctx context.Context, req usecase.TrialRequest) (*entity.Trial, error) {
	// Check if already has a trial
	existing, err := u.trialRepo.GetByClubID(ctx, req.ClubID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("club already has a trial")
	}

	// Get plan to know trial_days
	plan, err := u.planRepo.GetByKey(ctx, req.Plan)
	if err != nil || plan == nil {
		return nil, fmt.Errorf("plan not found: %s", req.Plan)
	}

	now := time.Now()
	trialEnd := now.AddDate(0, 0, plan.TrialDays)

	trial := &entity.Trial{
		ClubID:        req.ClubID,
		Plan:          req.Plan,
		Status:        entity.TrialStatusTrial,
		IsTrial:       true,
		TrialStartsAt: &now,
		TrialEndsAt:   &trialEnd,
		StartsAt:      now,
		EndsAt:        trialEnd,
	}

	if err := u.trialRepo.Upsert(ctx, trial); err != nil {
		return nil, fmt.Errorf("create trial: %w", err)
	}

	return trial, nil
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

	// 2. Distributed lock via Redis SetNX
	lockKey := fmt.Sprintf("checkout_lock:%s", req.IdempotencyKey)
	locked, err := u.redis.SetNX(ctx, lockKey, "1", 30*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lock: %w", err)
	}
	if !locked {
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

	// 3. Double-check DB after acquiring lock
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

	// 4. Get plan from DB
	plan, err := u.planRepo.GetByKey(ctx, req.Plan)
	if err != nil || plan == nil {
		return nil, fmt.Errorf("plan not found: %s", req.Plan)
	}

	externalRef := fmt.Sprintf("club_%s_plan_%s_%s", req.ClubID, req.Plan, req.IdempotencyKey)

	var (
		initPoint     string
		mpReferenceID string
		checkoutType  usecase.CheckoutType
	)

	if plan.MPPreapprovalPlanURL != "" {
		// ── Redirect flow: send user to MP's hosted subscription page ─────────
		// MP handles card collection. No card data touches our server.
		// The plan's init_point (stored at seed time) is the MP checkout URL.
		initPoint     = plan.MPPreapprovalPlanURL
		mpReferenceID = plan.MPPreapprovalPlanID
		checkoutType  = usecase.CheckoutTypePreapproval

		// Store intent in Redis so the webhook handler can find this club
		// when MP fires subscription_preapproval.updated (no external_reference
		// is set on the preapproval because the user subscribes via the plan's
		// generic init_point, not a per-checkout preapproval).
		if req.PayerEmail != "" {
			intentKey := fmt.Sprintf("sub_intent:%s", req.PayerEmail)
			intentVal := fmt.Sprintf("%s:%s", req.ClubID, req.Plan)
			_ = u.redis.Set(ctx, intentKey, intentVal, 24*time.Hour).Err()
		}
	} else {
		// ── One-time preference fallback (plan not yet seeded in MP) ──────────
		pref, err := u.mp.CreatePreference(ctx, MPPreferenceRequest{
			Items: []MPPreferenceItem{{
				Title:      fmt.Sprintf("NVF %s — Licencia mensual", plan.Name),
				Quantity:   1,
				UnitPrice:  plan.Price,
				CurrencyID: plan.Currency,
			}},
			ExternalReference: externalRef,
			NotificationURL:   u.mpCfg.NotificationURL,
		})
		if err != nil {
			return nil, fmt.Errorf("create MP preference: %w", err)
		}
		initPoint     = pref.InitPoint
		mpReferenceID = pref.ID
		checkoutType  = usecase.CheckoutTypePreference
	}

	// 6. Persist payment record
	payment := &entity.Payment{
		ClubID:         req.ClubID,
		UserID:         req.UserID,
		Type:           entity.PaymentTypeSubscription,
		Plan:           req.Plan,
		Amount:         plan.Price,
		Currency:       plan.Currency,
		Status:         entity.PaymentStatusPending,
		IdempotencyKey: req.IdempotencyKey,
		MPPreferenceID: mpReferenceID,
		InitPoint:      initPoint,
		ExternalRef:    externalRef,
	}
	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	return &usecase.CheckoutResponse{
		PaymentID:    payment.ID,
		InitPoint:    initPoint,
		CheckoutType: checkoutType,
		PreferenceID: func() string {
			if checkoutType == usecase.CheckoutTypePreference {
				return mpReferenceID
			}
			return ""
		}(),
	}, nil
}

func (u *paymentUsecaseImpl) HandleWebhook(ctx context.Context, payload usecase.WebhookPayload, xSignature, xRequestID string) error {
	dataID := payload.Data.ID
	if dataID == "" {
		return nil
	}

	// 1. Verify webhook signature
	if !u.mp.VerifyWebhookSignature(dataID, xRequestID, xSignature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// 2. Deduplicate — prevent reprocessing same notification
	dedupKey := fmt.Sprintf("webhook:%s:%s", dataID, payload.Action)
	set, err := u.redis.SetNX(ctx, dedupKey, "1", 5*time.Minute).Result()
	if err == nil && !set {
		return nil // already processed
	}

	action := payload.Action
	if action == "" {
		action = payload.Type
	}

	switch {
	case action == "payment.updated" || action == "payment.created":
		return u.handlePaymentEvent(ctx, dataID)

	case action == "subscription_preapproval.updated":
		return u.handlePreapprovalEvent(ctx, dataID)

	case action == "subscription_authorized_payment.updated":
		return u.handleAuthorizedPaymentEvent(ctx, dataID)

	default:
		// Unrecognised event — ignore safely
		return nil
	}
}

// handlePaymentEvent processes a one-time checkout payment event.
func (u *paymentUsecaseImpl) handlePaymentEvent(ctx context.Context, mpPaymentIDStr string) error {
	mpPaymentID, err := strconv.ParseInt(mpPaymentIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid mp payment id %q: %w", mpPaymentIDStr, err)
	}

	mpResp, err := u.mp.GetPayment(ctx, mpPaymentID)
	if err != nil {
		return fmt.Errorf("fetch MP payment: %w", err)
	}

	// Locate our internal record — first by MP payment ID, fallback to external_reference
	payment, err := u.paymentRepo.GetByMPPaymentID(ctx, mpPaymentID)
	if err != nil {
		return fmt.Errorf("find payment by mp_id: %w", err)
	}
	if payment == nil && mpResp.ExternalReference != "" {
		payment, err = u.paymentRepo.GetByExternalRef(ctx, mpResp.ExternalReference)
		if err != nil {
			return fmt.Errorf("find payment by ext_ref: %w", err)
		}
	}
	if payment == nil {
		return nil // not our payment
	}

	newStatus := mapMPPaymentStatus(mpResp.Status)
	if err := u.paymentRepo.UpdateStatus(ctx, payment.ID, newStatus, &mpPaymentID); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	if newStatus == entity.PaymentStatusApproved && payment.Type == entity.PaymentTypeSubscription {
		return u.activateSubscription(ctx, payment.ClubID, payment.Plan, &payment.ID)
	}

	if (newStatus == entity.PaymentStatusRejected || newStatus == entity.PaymentStatusCanceled) &&
		payment.Type == entity.PaymentTypeSubscription {
		return u.expireSubscription(ctx, payment.ClubID)
	}

	// Activate marketplace ad after confirmed payment
	if newStatus == entity.PaymentStatusApproved && payment.Type == entity.PaymentTypeAdvertising {
		if payment.ItemID != nil {
			u.activateMarketplaceAd(ctx, *payment.ItemID)
		}
	}

	return nil
}

// handlePreapprovalEvent processes a subscription plan status change.
// Triggered when a subscription is paused, cancelled or reactivated in MP.
func (u *paymentUsecaseImpl) handlePreapprovalEvent(ctx context.Context, preapprovalID string) error {
	preapproval, err := u.mp.GetPreapproval(ctx, preapprovalID)
	if err != nil {
		return fmt.Errorf("fetch preapproval: %w", err)
	}

	// Locate club via external_reference (set on one-time preference checkouts).
	payment, err := u.paymentRepo.GetByExternalRef(ctx, preapproval.ExternalReference)
	if err != nil {
		return fmt.Errorf("find payment by ext_ref: %w", err)
	}

	var clubID, plan string

	if payment != nil {
		clubID = payment.ClubID
		plan = payment.Plan
	} else {
		// Redirect flow: no external_reference was set because the user subscribed
		// via the plan's generic init_point. Fall back to the intent we stored in
		// Redis at checkout time, keyed by payer email.
		if preapproval.PayerEmail == "" {
			return nil // can't identify club — ignore safely
		}
		intentKey := fmt.Sprintf("sub_intent:%s", preapproval.PayerEmail)
		intentVal, redisErr := u.redis.Get(ctx, intentKey).Result()
		if redisErr != nil || intentVal == "" {
			return nil // intent expired or never set — ignore
		}
		parts := splitN(intentVal, ":", 2)
		if len(parts) != 2 {
			return nil
		}
		clubID, plan = parts[0], parts[1]
		// Intent consumed — clean up
		_ = u.redis.Del(ctx, intentKey).Err()
	}

	trial, err := u.trialRepo.GetByClubID(ctx, clubID)
	if err != nil {
		return fmt.Errorf("get trial: %w", err)
	}

	switch preapproval.Status {
	case "authorized":
		if trial == nil || trial.Status != entity.TrialStatusActive {
			if activateErr := u.activateSubscription(ctx, clubID, plan, nil); activateErr != nil {
				return fmt.Errorf("activate subscription: %w", activateErr)
			}
			// Refresh trial for UpdatePreapproval below
			trial, _ = u.trialRepo.GetByClubID(ctx, clubID)
		}
	case "paused", "cancelled":
		return u.expireSubscription(ctx, clubID)
	}

	// Persist latest preapproval ID and status for traceability.
	if trial != nil {
		_ = u.trialRepo.UpdatePreapproval(ctx, trial.ID, preapproval.ID, preapproval.Status)
	}

	return nil
}

// handleAuthorizedPaymentEvent processes a recurring subscription payment.
// Triggered when MP automatically charges the monthly fee.
func (u *paymentUsecaseImpl) handleAuthorizedPaymentEvent(ctx context.Context, authorizedPaymentID string) error {
	apID, err := strconv.ParseInt(authorizedPaymentID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid authorized payment id %q: %w", authorizedPaymentID, err)
	}

	authPayment, err := u.mp.GetAuthorizedPayment(ctx, apID)
	if err != nil {
		return fmt.Errorf("fetch authorized payment: %w", err)
	}

	if authPayment.Status != "processed" {
		return nil // recycling or cancelled — no action needed yet
	}

	// Find existing subscription via preapproval external_reference.
	// We look up the preapproval to get the external_reference we set at checkout time.
	if authPayment.PreapprovalID == "" {
		return nil
	}

	preapproval, err := u.mp.GetPreapproval(ctx, authPayment.PreapprovalID)
	if err != nil || preapproval == nil {
		return nil
	}

	// Try to locate club via external_reference first (one-time preference checkouts).
	var clubID string
	payment, err := u.paymentRepo.GetByExternalRef(ctx, preapproval.ExternalReference)
	if err != nil {
		return fmt.Errorf("find payment for authorized payment: %w", err)
	}
	if payment != nil {
		clubID = payment.ClubID
	} else if preapproval.PayerEmail != "" {
		// Redirect flow fallback: look up Redis intent by payer email.
		intentKey := fmt.Sprintf("sub_intent:%s", preapproval.PayerEmail)
		intentVal, _ := u.redis.Get(ctx, intentKey).Result()
		if intentVal != "" {
			parts := splitN(intentVal, ":", 2)
			if len(parts) == 2 {
				clubID = parts[0]
			}
		}
	}

	if clubID == "" {
		return nil
	}

	// Renew by extending EndsAt 1 more month
	trial, err := u.trialRepo.GetByClubID(ctx, clubID)
	if err != nil || trial == nil {
		return nil
	}

	now := time.Now()
	newEnd := trial.EndsAt.AddDate(0, 1, 0)
	if newEnd.Before(now) {
		newEnd = now.AddDate(0, 1, 0)
	}
	renewed := &entity.Trial{
		ClubID:    clubID,
		Plan:      trial.Plan,
		Status:    entity.TrialStatusActive,
		IsTrial:   false,
		PaymentID: trial.PaymentID,
		StartsAt:  trial.StartsAt,
		EndsAt:    newEnd,
	}
	return u.trialRepo.Upsert(ctx, renewed)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (u *paymentUsecaseImpl) activateSubscription(ctx context.Context, clubID, plan string, paymentID *uint) error {
	now := time.Now()
	trial := &entity.Trial{
		ClubID:    clubID,
		Plan:      plan,
		Status:    entity.TrialStatusActive,
		IsTrial:   false,
		PaymentID: paymentID,
		StartsAt:  now,
		EndsAt:    now.AddDate(0, 1, 0),
	}
	return u.trialRepo.Upsert(ctx, trial)
}

func (u *paymentUsecaseImpl) expireSubscription(ctx context.Context, clubID string) error {
	trial, err := u.trialRepo.GetByClubID(ctx, clubID)
	if err != nil || trial == nil {
		return err
	}
	if trial.Status == entity.TrialStatusExpired {
		return nil
	}
	return u.trialRepo.UpdateStatus(ctx, trial.ID, entity.TrialStatusExpired)
}

// splitN is a thin wrapper around strings.SplitN for readability.
func splitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

// activateMarketplaceAd publishes an event to RabbitMQ so sports-ms can activate the ad.
// Errors are logged but not propagated — the payment is already confirmed.
func (u *paymentUsecaseImpl) activateMarketplaceAd(ctx context.Context, itemID int64) {
	if u.adPublisher == nil {
		fmt.Printf("WARN: RabbitMQ publisher not available; ad %d will not be activated\n", itemID)
		return
	}
	payload, err := json.Marshal(map[string]int64{"item_id": itemID})
	if err != nil {
		return
	}
	if err := u.adPublisher.PublishMessageWithCtx(ctx, payload); err != nil {
		fmt.Printf("ERROR: failed to publish ad.approved event for item %d: %v\n", itemID, err)
	}
}

func (u *paymentUsecaseImpl) ListPayments(ctx context.Context, from, to time.Time) ([]entity.Payment, error) {
	return u.paymentRepo.ListByDateRange(ctx, from, to)
}

func (u *paymentUsecaseImpl) RegisterManualPayment(ctx context.Context, paymentID uint, req usecase.ManualPaymentRequest) error {
	payment, err := u.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("get payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found")
	}
	if payment.Status != entity.PaymentStatusPending {
		return fmt.Errorf("payment is not pending (current status: %s)", payment.Status)
	}

	if err := u.paymentRepo.UpdateStatus(ctx, payment.ID, entity.PaymentStatusApproved, nil); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	if payment.Type == entity.PaymentTypeSubscription {
		return u.activateSubscription(ctx, payment.ClubID, payment.Plan, &payment.ID)
	}

	if payment.Type == entity.PaymentTypeAdvertising && payment.ItemID != nil {
		u.activateMarketplaceAd(ctx, *payment.ItemID)
	}

	return nil
}

func mapMPPaymentStatus(mpStatus string) entity.PaymentStatus {
	switch mpStatus {
	case "approved":
		return entity.PaymentStatusApproved
	case "rejected":
		return entity.PaymentStatusRejected
	case "cancelled":
		return entity.PaymentStatusCanceled
	default:
		return entity.PaymentStatusPending
	}
}

// GetMarketplaceCommission returns the active commission config for marketplace listings.
func (u *paymentUsecaseImpl) GetMarketplaceCommission(ctx context.Context) (*usecase.MarketplaceCommissionResponse, error) {
	commission, err := u.commissionRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("get marketplace commission: %w", err)
	}
	if commission == nil {
		// Fallback defaults so the frontend never gets a 500
		return &usecase.MarketplaceCommissionResponse{
			Percentage: 0.02,
			WordingES:  "Se cobra una comisión del 2% sobre el precio de venta al publicar.",
			WordingEN:  "A 2% commission on the sale price is charged upon listing.",
			Currency:   "ARS",
		}, nil
	}
	return &usecase.MarketplaceCommissionResponse{
		Percentage: commission.Percentage,
		WordingES:  commission.WordingES,
		WordingEN:  commission.WordingEN,
		Currency:   commission.Currency,
	}, nil
}

// CreateMarketplaceCheckout creates a MercadoPago preference for the advertising commission.
func (u *paymentUsecaseImpl) CreateMarketplaceCheckout(ctx context.Context, req usecase.MarketplaceCheckoutRequest) (*usecase.MarketplaceCheckoutResponse, error) {
	// 1. Idempotency check
	existing, err := u.paymentRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("idempotency check: %w", err)
	}
	if existing != nil {
		return &usecase.MarketplaceCheckoutResponse{
			PaymentID:     existing.ID,
			CommissionAmt: existing.Amount,
			InitPoint:     existing.InitPoint,
			PreferenceID:  existing.MPPreferenceID,
		}, nil
	}

	// 2. Get active commission config
	commission, err := u.commissionRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("get commission config: %w", err)
	}
	percentage := 0.02
	currency := "ARS"
	if commission != nil {
		percentage = commission.Percentage
		currency = commission.Currency
	}

	commissionAmt := req.Price * percentage
	if commissionAmt < 1 {
		commissionAmt = 1 // MP minimum
	}

	externalRef := fmt.Sprintf("marketplace_user_%s_%s", req.UserID, req.IdempotencyKey)

	// Build back URLs: base is mpCfg.BackURL, suffixed with payment result
	baseURL := u.mpCfg.BackURL
	backURLs := &MPBackURLs{
		Success: baseURL + "?payment=success&ref=" + externalRef,
		Pending: baseURL + "?payment=pending&ref=" + externalRef,
		Failure: baseURL + "?payment=failure&ref=" + externalRef,
	}

	// 3. Create MP Checkout Pro preference (single payment, no split)
	pref, err := u.mp.CreatePreference(ctx, MPPreferenceRequest{
		Items: []MPPreferenceItem{{
			Title:      fmt.Sprintf("Comisión publicación: %s", req.ItemTitle),
			Quantity:   1,
			UnitPrice:  commissionAmt,
			CurrencyID: currency,
		}},
		ExternalReference: externalRef,
		NotificationURL:   u.mpCfg.NotificationURL,
		BackURLs:          backURLs,
		AutoReturn:        "approved",
	})
	if err != nil {
		return nil, fmt.Errorf("create MP preference: %w", err)
	}

	// 4. Persist payment record
	payment := &entity.Payment{
		ClubID:         req.UserID, // reuse club_id field for user_id context
		UserID:         req.UserID,
		Type:           entity.PaymentTypeAdvertising,
		Amount:         commissionAmt,
		Currency:       currency,
		Status:         entity.PaymentStatusPending,
		IdempotencyKey: req.IdempotencyKey,
		MPPreferenceID: pref.ID,
		InitPoint:      pref.InitPoint,
		ExternalRef:    externalRef,
		ItemID:         &req.ItemID,
	}
	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("save marketplace payment: %w", err)
	}

	return &usecase.MarketplaceCheckoutResponse{
		PaymentID:     payment.ID,
		CommissionAmt: commissionAmt,
		InitPoint:     pref.InitPoint,
		PreferenceID:  pref.ID,
	}, nil
}
