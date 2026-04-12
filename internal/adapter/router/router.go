package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jnieto01/payments-ms-01/internal/adapter/rest/handler"
	"github.com/jnieto01/payments-ms-01/internal/middleware"
	utilsjwt "github.com/jnieto01/utils-01/jwt"
)

func SetupRouter(paymentHandler *handler.PaymentHandler, jwtService utilsjwt.JWTService) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group("/payments-ms-01/api")
	{
		// Public endpoints
		api.POST("/payments/webhook", paymentHandler.HandleWebhook)
		api.GET("/plans", paymentHandler.GetPlans)
		api.GET("/subscriptions/status/:clubId", paymentHandler.GetSubscriptionStatus)
		api.GET("/marketplace/commission", paymentHandler.GetMarketplaceCommission)

		// Protected endpoints
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			protected.POST("/subscriptions/checkout", paymentHandler.CreateSubscriptionCheckout)
			protected.POST("/subscriptions/trial", paymentHandler.StartTrial)
			protected.POST("/marketplace/checkout", paymentHandler.CreateMarketplaceCheckout)
		}
	}

	return r
}
