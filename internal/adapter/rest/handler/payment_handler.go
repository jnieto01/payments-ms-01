package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jnieto01/payments-ms-01/internal/usecase"
)

type PaymentHandler struct {
	uc usecase.PaymentUsecase
}

func NewPaymentHandler(uc usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

// GET /plans
func (h *PaymentHandler) GetPlans(c *gin.Context) {
	plans, err := h.uc.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plans})
}

// GET /subscriptions/status/:clubId
func (h *PaymentHandler) GetSubscriptionStatus(c *gin.Context) {
	clubID := c.Param("clubId")
	if clubID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "club_id required"})
		return
	}

	status, err := h.uc.GetTrialStatus(c.Request.Context(), clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": status})
}

// POST /subscriptions/trial
func (h *PaymentHandler) StartTrial(c *gin.Context) {
	var req usecase.TrialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.uc.StartTrial(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": sub})
}

// POST /subscriptions/checkout
func (h *PaymentHandler) CreateSubscriptionCheckout(c *gin.Context) {
	var req usecase.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enrich request from JWT context — never trust these fields from the body.
	if userID, ok := c.Get("userId"); ok && userID != nil {
		req.UserID = userID.(string)
	}


	if email, ok := c.Get("email"); ok && email != nil {
		req.PayerEmail = email.(string)
	}

	println("paso user_email")



	resp, err := h.uc.CreateSubscriptionCheckout(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GET /marketplace/commission
func (h *PaymentHandler) GetMarketplaceCommission(c *gin.Context) {
	commission, err := h.uc.GetMarketplaceCommission(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": commission})
}

// POST /marketplace/checkout
func (h *PaymentHandler) CreateMarketplaceCheckout(c *gin.Context) {
	var req usecase.MarketplaceCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userID, ok := c.Get("userId"); ok && userID != nil {
		req.UserID = userID.(string)
	}
	if email, ok := c.Get("email"); ok && email != nil {
		req.PayerEmail = email.(string)
	}

	resp, err := h.uc.CreateMarketplaceCheckout(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GET /admin/payments?from=YYYY-MM-DD&to=YYYY-MM-DD
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	today := time.Now().Format("2006-01-02")
	fromStr := c.DefaultQuery("from", today)
	toStr := c.DefaultQuery("to", today)

	from, err := time.ParseInLocation("2006-01-02", fromStr, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date, expected YYYY-MM-DD"})
		return
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date, expected YYYY-MM-DD"})
		return
	}
	// include the full last day
	to = to.Add(24*time.Hour - time.Second)

	payments, err := h.uc.ListPayments(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": payments})
}

// POST /admin/payments/:id/manual
func (h *PaymentHandler) RegisterManualPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
		return
	}

	var req usecase.ManualPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.RegisterManualPayment(c.Request.Context(), uint(id), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

// POST /payments/webhook
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {

	var payload usecase.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signature := c.GetHeader("X-Signature")
	requestID := c.GetHeader("X-Request-ID")
	if err := h.uc.HandleWebhook(c.Request.Context(), payload, signature, requestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
