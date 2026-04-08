package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jnieto01/payments-ms/internal/config"
)

type mpClient struct {
	cfg        config.MercadoPagoConfig
	httpClient *http.Client
}

func newMPClient(cfg config.MercadoPagoConfig) *mpClient {
	return &mpClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type MPPreferenceItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	CurrencyID  string  `json:"currency_id"`
}

type MPPreferenceRequest struct {
	Items              []MPPreferenceItem `json:"items"`
	ExternalReference  string             `json:"external_reference"`
	NotificationURL    string             `json:"notification_url,omitempty"`
	BackURLs           *MPBackURLs        `json:"back_urls,omitempty"`
	AutoReturn         string             `json:"auto_return,omitempty"`
}

type MPBackURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}

type MPPreferenceResponse struct {
	ID        string `json:"id"`
	InitPoint string `json:"init_point"`
}

type MPPaymentResponse struct {
	ID             int64   `json:"id"`
	Status         string  `json:"status"`
	StatusDetail   string  `json:"status_detail"`
	ExternalRef    string  `json:"external_reference"`
	TransactionAmt float64 `json:"transaction_amount"`
}

func (c *mpClient) CreatePreference(ctx context.Context, req MPPreferenceRequest) (*MPPreferenceResponse, error) {
	url := fmt.Sprintf("%s/checkout/preferences", c.cfg.BaseURL)
	return doPost[MPPreferenceRequest, MPPreferenceResponse](ctx, c, url, req)
}

func (c *mpClient) GetPayment(ctx context.Context, paymentID string) (*MPPaymentResponse, error) {
	url := fmt.Sprintf("%s/v1/payments/%s", c.cfg.BaseURL, paymentID)
	return doGet[MPPaymentResponse](ctx, c, url)
}

func doPost[Req any, Res any](ctx context.Context, c *mpClient, url string, body Req) (*Res, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("mercadopago error %d: %s", resp.StatusCode, string(raw))
	}

	var result Res
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func doGet[Res any](ctx context.Context, c *mpClient, url string) (*Res, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("mercadopago error %d: %s", resp.StatusCode, string(raw))
	}

	var result Res
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
