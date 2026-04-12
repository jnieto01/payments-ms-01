package impl

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	sdkcfg           "github.com/mercadopago/sdk-go/pkg/config"
	sdkpayment       "github.com/mercadopago/sdk-go/pkg/payment"
	sdkpreapproval   "github.com/mercadopago/sdk-go/pkg/preapproval"
	sdkplan          "github.com/mercadopago/sdk-go/pkg/preapprovalplan"
	sdkpreference    "github.com/mercadopago/sdk-go/pkg/preference"

	appconfig "github.com/jnieto01/payments-ms-01/internal/config"
)

// mpClient wraps the official MercadoPago SDK clients.
// Webhook signature verification is kept custom: the SDK does not cover it.
type mpClient struct {
	webhookSecret      string
	preferenceCli      sdkpreference.Client
	paymentCli         sdkpayment.Client
	preapprovalCli     sdkpreapproval.Client
	preapprovalPlanCli sdkplan.Client
}

func newMPClient(cfg appconfig.MercadoPagoConfig) *mpClient {
	c, err := sdkcfg.New(cfg.AccessToken)
	if err != nil {
		panic(fmt.Sprintf("failed to init MercadoPago SDK: %v", err))
	}
	return &mpClient{
		webhookSecret:      cfg.WebhookSecret,
		preferenceCli:      sdkpreference.NewClient(c),
		paymentCli:         sdkpayment.NewClient(c),
		preapprovalCli:     sdkpreapproval.NewClient(c),
		preapprovalPlanCli: sdkplan.NewClient(c),
	}
}

// ── Request / Response types ──────────────────────────────────────────────────
// Internal types used by the usecase layer. They map to SDK types internally
// so the usecase layer stays decoupled from the SDK package.

type MPPreferenceItem struct {
	Title      string
	Quantity   int
	UnitPrice  float64
	CurrencyID string
}

type MPBackURLs struct {
	Success string
	Pending string
	Failure string
}

type MPPreferenceRequest struct {
	Items             []MPPreferenceItem
	ExternalReference string
	NotificationURL   string
	BackURLs          *MPBackURLs
	AutoReturn        string // "approved" | "all" — MP redirects automatically after payment
}

type MPPreferenceResponse struct {
	ID        string
	InitPoint string
}

type MPPaymentResponse struct {
	ID                int64
	Status            string
	ExternalReference string
	TransactionAmount float64
}

type MPPreapprovalResponse struct {
	ID                string
	Status            string
	ExternalReference string
	PayerEmail        string
}

type MPAuthorizedPaymentResponse struct {
	ID            int64
	Status        string
	PreapprovalID string
}

type MPPreapprovalPlanRequest struct {
	Reason        string
	Frequency     int
	FrequencyType string
	Amount        float64
	CurrencyID    string
	BackURL       string
}

type MPPreapprovalPlanResponse struct {
	ID        string
	InitPoint string
}

// ── Methods ──────────────────────────────────────────────────────────────────

func (c *mpClient) CreatePreference(ctx context.Context, req MPPreferenceRequest) (*MPPreferenceResponse, error) {
	items := make([]sdkpreference.ItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = sdkpreference.ItemRequest{
			Title:      item.Title,
			Quantity:   item.Quantity,
			UnitPrice:  item.UnitPrice,
			CurrencyID: item.CurrencyID,
		}
	}
	sdkReq := sdkpreference.Request{
		Items:             items,
		ExternalReference: req.ExternalReference,
		NotificationURL:   req.NotificationURL,
		AutoReturn:        req.AutoReturn,
	}
	if req.BackURLs != nil {
		sdkReq.BackURLs = &sdkpreference.BackURLsRequest{
			Success: req.BackURLs.Success,
			Pending: req.BackURLs.Pending,
			Failure: req.BackURLs.Failure,
		}
	}
	resp, err := c.preferenceCli.Create(ctx, sdkReq)
	if err != nil {
		return nil, fmt.Errorf("MP CreatePreference: %w", err)
	}
	return &MPPreferenceResponse{ID: resp.ID, InitPoint: resp.InitPoint}, nil
}

func (c *mpClient) GetPayment(ctx context.Context, paymentID int64) (*MPPaymentResponse, error) {
	resp, err := c.paymentCli.Get(ctx, int(paymentID))
	if err != nil {
		return nil, fmt.Errorf("MP GetPayment(%d): %w", paymentID, err)
	}
	return &MPPaymentResponse{
		ID:                int64(resp.ID),
		Status:            string(resp.Status),
		ExternalReference: resp.ExternalReference,
		TransactionAmount: resp.TransactionAmount,
	}, nil
}

func (c *mpClient) GetPreapproval(ctx context.Context, preapprovalID string) (*MPPreapprovalResponse, error) {
	resp, err := c.preapprovalCli.Get(ctx, preapprovalID)
	if err != nil {
		return nil, fmt.Errorf("MP GetPreapproval(%s): %w", preapprovalID, err)
	}
	return &MPPreapprovalResponse{
		ID:                resp.ID,
		Status:            resp.Status,
		ExternalReference: resp.ExternalReference,
		PayerEmail:        resp.PayerEmail,
	}, nil
}

// GetAuthorizedPayment fetches a recurring authorized payment via the payments endpoint.
// MP authorized payments are returned as regular payment objects.
// The preapproval ID (subscription reference) is in PointOfInteraction.TransactionData.SubscriptionID.
func (c *mpClient) GetAuthorizedPayment(ctx context.Context, authorizedPaymentID int64) (*MPAuthorizedPaymentResponse, error) {
	resp, err := c.paymentCli.Get(ctx, int(authorizedPaymentID))
	if err != nil {
		return nil, fmt.Errorf("MP GetAuthorizedPayment(%d): %w", authorizedPaymentID, err)
	}
	return &MPAuthorizedPaymentResponse{
		ID:            int64(resp.ID),
		Status:        string(resp.Status),
		PreapprovalID: resp.PointOfInteraction.TransactionData.SubscriptionID,
	}, nil
}

// CreatePreapprovalPlan creates a recurring subscription plan in MercadoPago.
// Called once per plan via the seed script — not at request time.
func (c *mpClient) CreatePreapprovalPlan(ctx context.Context, req MPPreapprovalPlanRequest) (*MPPreapprovalPlanResponse, error) {
	sdkReq := sdkplan.Request{
		Reason:  req.Reason,
		BackURL: req.BackURL,
		AutoRecurring: &sdkplan.AutoRecurringRequest{
			Frequency:         req.Frequency,
			FrequencyType:     req.FrequencyType,
			TransactionAmount: req.Amount,
			CurrencyID:        req.CurrencyID,
		},
	}
	resp, err := c.preapprovalPlanCli.Create(ctx, sdkReq)
	if err != nil {
		return nil, fmt.Errorf("MP CreatePreapprovalPlan: %w", err)
	}
	return &MPPreapprovalPlanResponse{ID: resp.ID, InitPoint: resp.InitPoint}, nil
}

// VerifyWebhookSignature validates the X-Signature header sent by MercadoPago.
// The SDK does not cover this — kept as custom HMAC-SHA256 verification.
// MP format: "ts=<timestamp>,v1=<hmac>"
// Signed payload: "id:<data.id>;request-id:<x-request-id>;ts:<ts>;"
func (c *mpClient) VerifyWebhookSignature(dataID, requestID, xSignature string) bool {
	if c.webhookSecret == "" || xSignature == "" {
		return true // skip verification if secret not configured
	}

	var ts, v1 string
	for _, part := range strings.Split(xSignature, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "ts":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}

	if ts == "" || v1 == "" {
		return false
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, requestID, ts)
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(v1))
}
