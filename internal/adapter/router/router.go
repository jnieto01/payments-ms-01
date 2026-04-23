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

		// Protected endpoints (any authenticated user)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtService))
		{
			protected.POST("/subscriptions/checkout", paymentHandler.CreateSubscriptionCheckout)
			protected.POST("/subscriptions/trial", paymentHandler.StartTrial)
			protected.POST("/marketplace/checkout", paymentHandler.CreateMarketplaceCheckout)
		}

		// Admin-only endpoints (role: admin_nvf)
		admin := api.Group("/admin")
		admin.Use(middleware.AdminMiddleware(jwtService))
		{
			admin.GET("/payments", paymentHandler.ListPayments)
			admin.POST("/payments/:id/manual", paymentHandler.RegisterManualPayment)
			admin.GET("/trials", paymentHandler.AdminListTrials)
			admin.POST("/trials", paymentHandler.AdminAssignTrial)
			admin.PATCH("/trials/:id", paymentHandler.AdminExtendTrial)
			admin.DELETE("/trials/:id", paymentHandler.AdminCancelTrial)
			admin.GET("/plans", paymentHandler.AdminListPlans)
			admin.GET("/plans/:id", paymentHandler.AdminGetPlanByID)
			admin.POST("/plans", paymentHandler.AdminCreatePlan)
			admin.PUT("/plans/:id", paymentHandler.AdminUpdatePlan)
			admin.PATCH("/plans/:id/toggle", paymentHandler.AdminTogglePlan)
			admin.POST("/plans/:id/mp-link", paymentHandler.AdminLinkMPPlan)
		}
	}

	return r
}
