package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jnieto01/payments-ms-01/internal/usecase"
)

var validate = validator.New()

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	if err := validate.Struct(&req); err != nil {
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
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enrich request from JWT context — never trust these fields from the body.
	if userID, ok := c.Get("user_id"); ok && userID != nil {
		req.UserID = userID.(string)
	}
	if email, ok := c.Get("user_email"); ok && email != nil {
		req.PayerEmail = email.(string)
	}

	resp, err := h.uc.CreateSubscriptionCheckout(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GET /marketplace/commission
func (h *PaymentHandler) GetMarketplaceCommission(c *gin.Context) {
	commission, err := h.uc.GetMarketplaceCommission(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if userID, ok := c.Get("user_id"); ok && userID != nil {
		req.UserID = userID.(string)
	}
	if email, ok := c.Get("user_email"); ok && email != nil {
		req.PayerEmail = email.(string)
	}

	resp, err := h.uc.CreateMarketplaceCheckout(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.RegisterManualPayment(c.Request.Context(), uint(id), req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

// GET /admin/trials
func (h *PaymentHandler) AdminListTrials(c *gin.Context) {
	trials, err := h.uc.AdminListTrials(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trials})
}

// POST /admin/trials
func (h *PaymentHandler) AdminAssignTrial(c *gin.Context) {
	var req usecase.AdminAssignTrialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.AdminAssignTrial(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "assigned"})
}

// PATCH /admin/trials/:id
func (h *PaymentHandler) AdminExtendTrial(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trial id"})
		return
	}
	var req usecase.AdminExtendTrialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.AdminExtendTrial(c.Request.Context(), id, req.Days); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "extended"})
}

// DELETE /admin/trials/:id
func (h *PaymentHandler) AdminCancelTrial(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid trial id"})
		return
	}
	if err := h.uc.AdminCancelTrial(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}

// GET /admin/plans
func (h *PaymentHandler) AdminListPlans(c *gin.Context) {
	plans, err := h.uc.AdminGetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plans})
}

// GET /admin/plans/:id
func (h *PaymentHandler) AdminGetPlanByID(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	plan, err := h.uc.AdminGetPlanByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if plan == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// POST /admin/plans
func (h *PaymentHandler) AdminCreatePlan(c *gin.Context) {
	var req usecase.CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.uc.AdminCreatePlan(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": plan})
}

// PUT /admin/plans/:id
func (h *PaymentHandler) AdminUpdatePlan(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	var req usecase.UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	plan, err := h.uc.AdminUpdatePlan(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": plan})
}

// PATCH /admin/plans/:id/toggle
func (h *PaymentHandler) AdminTogglePlan(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	if err := h.uc.AdminTogglePlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// POST /admin/plans/:id/mp-link
func (h *PaymentHandler) AdminLinkMPPlan(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	if err := h.uc.AdminLinkMPPlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "linked"})
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	return uint(v), err
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
