package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jnieto01/payments-ms/internal/adapter/rest/handler"
	"github.com/jnieto01/payments-ms/internal/middleware"
)

func SetupRouter(paymentHandler *handler.PaymentHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/payments-ms/api")
	{
		// Public — webhook from MercadoPago (no auth required)
		api.POST("/payments/webhook", paymentHandler.HandleWebhook)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/subscriptions/checkout", paymentHandler.CreateSubscriptionCheckout)
			protected.GET("/subscriptions/club/:clubId", paymentHandler.GetSubscription)
		}
	}

	return r
}
