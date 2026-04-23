package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jnieto01/payments-ms-01/internal/adapter/rest/handler"
	"github.com/jnieto01/payments-ms-01/internal/adapter/router"
	appconfig "github.com/jnieto01/payments-ms-01/internal/config"
	repoImpl "github.com/jnieto01/payments-ms-01/internal/repository"
	"github.com/jnieto01/payments-ms-01/internal/usecase"
	"github.com/jnieto01/payments-ms-01/internal/usecase/impl"
	"github.com/jnieto01/utils-01/database"
	utilsjwt "github.com/jnieto01/utils-01/jwt"
	"github.com/jnieto01/utils-01/migration"
	"github.com/jnieto01/utils-01/rabbitmq"
	"github.com/redis/go-redis/v9"
)

const pathConfig = "./config/"

func main() {
	cfg := appconfig.Load(pathConfig)

	// Database
	mysqlDB, err := database.NewDatabase(cfg.MySQL)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	// Migrations
	if err := migration.RunAutoMigrations(&cfg.MySQL, cfg.MySQLMigration); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Redis (raw client for SetNX idempotency)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		PoolSize: cfg.RedisPoolSize,
	})

	// JWT service — same implementation as user-ms for compatible token validation
	jwtService := utilsjwt.NewJWTService(cfg.JWT.SecretKey)

	// Wire up dependencies
	paymentRepo := repoImpl.NewPaymentRepository(mysqlDB.DB)
	trialRepo := repoImpl.NewTrialRepository(mysqlDB.DB)
	planRepo := repoImpl.NewPlanRepository(mysqlDB.DB)
	commissionRepo := repoImpl.NewMarketplaceCommissionRepository(mysqlDB.DB)
	paymentUC := impl.NewPaymentUsecase(paymentRepo, trialRepo, planRepo, commissionRepo, redisClient, cfg.MercadoPago, cfg.RabbitMQ)
	paymentHandler := handler.NewPaymentHandler(paymentUC)

	// Start RabbitMQ consumer for manual advertising payment confirmations from sports-ms
	if cfg.RabbitMQ.Host != "" {
		go startManualAdPaymentConsumer(cfg.RabbitMQ, paymentUC)
	}

	r := router.SetupRouter(paymentHandler, jwtService)

	log.Printf("payments-ms-01 starting on port %d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func startManualAdPaymentConsumer(rmqCfg rabbitmq.Config, uc usecase.PaymentUsecase) {
	consumer, err := rabbitmq.NewMessageService(rmqCfg, "advertising.payment.manual.confirmed")
	if err != nil {
		log.Printf("WARN: manual-ad-payment consumer: failed to connect to RabbitMQ: %v", err)
		return
	}
	log.Printf("manual-ad-payment consumer: listening on queue advertising.payment.manual.confirmed")
	if err := consumer.Consume(context.Background(), "payments-ms-ad-manual-consumer", func(body []byte) error {
		var msg usecase.AdvertisingPaymentConfirmedMsg
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("invalid advertising.payment.manual.confirmed payload: %w", err)
		}
		return uc.HandleManualAdPaymentConfirmed(context.Background(), msg)
	}); err != nil {
		log.Printf("ERROR: manual-ad-payment consumer: %v", err)
	}
}
