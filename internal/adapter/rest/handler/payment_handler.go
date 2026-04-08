package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jnieto01/payments-ms/internal/usecase"
)

type PaymentHandler struct {
	uc usecase.PaymentUsecase
}

func NewPaymentHandler(uc usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) CreateSubscriptionCheckout(c *gin.Context) {
	var req usecase.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract user ID from JWT claims (set by auth middleware)
	userID, _ := c.Get("user_id")
	if userID != nil {
		req.UserID = userID.(string)
	}

	resp, err := h.uc.CreateSubscriptionCheckout(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	var payload usecase.WebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signature := c.GetHeader("X-Signature")
	if err := h.uc.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *PaymentHandler) GetSubscription(c *gin.Context) {
	clubID := c.Param("clubId")
	if clubID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "club_id required"})
		return
	}

	sub, err := h.uc.GetSubscriptionByClub(c.Request.Context(), clubID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub == nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sub})
}
