package main

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jnieto01/payments-ms/internal/adapter/rest/handler"
	"github.com/jnieto01/payments-ms/internal/adapter/router"
	appconfig "github.com/jnieto01/payments-ms/internal/config"
	repoImpl "github.com/jnieto01/payments-ms/internal/repository"
	"github.com/jnieto01/payments-ms/internal/usecase/impl"
	"github.com/pressly/goose/v3"
	"github.com/spf13/viper"
)

func main() {
	appconfig.LoadConfig()

	db := appconfig.InitDB()
	redisClient := appconfig.InitRedis()
	mpCfg := appconfig.LoadMercadoPagoConfig()

	// Run migrations
	if viper.GetBool("golang-migrate.mysql.enabled") {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("failed to get sql.DB: %v", err)
		}
		if err := goose.SetDialect("mysql"); err != nil {
			log.Fatalf("goose set dialect: %v", err)
		}
		migrationsDir := viper.GetString("golang-migrate.mysql.locations")
		if err := goose.Up(sqlDB, migrationsDir); err != nil {
			log.Fatalf("goose up: %v", err)
		}
		log.Println("migrations applied successfully")
	}

	// Wire up dependencies
	paymentRepo := repoImpl.NewPaymentRepository(db)
	subscriptionRepo := repoImpl.NewSubscriptionRepository(db)
	paymentUC := impl.NewPaymentUsecase(paymentRepo, subscriptionRepo, redisClient, mpCfg)
	paymentHandler := handler.NewPaymentHandler(paymentUC)

	r := router.SetupRouter(paymentHandler)

	port := viper.GetInt("server.port")
	log.Printf("payments-ms starting on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
