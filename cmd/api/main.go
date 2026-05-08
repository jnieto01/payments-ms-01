package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jnieto01/payments-ms-01/internal/adapter/rest/handler"
	"github.com/jnieto01/payments-ms-01/internal/adapter/router"
	appconfig "github.com/jnieto01/payments-ms-01/internal/config"
	repoImpl "github.com/jnieto01/payments-ms-01/internal/repository"
	"github.com/jnieto01/payments-ms-01/internal/usecase"
	"github.com/jnieto01/payments-ms-01/internal/usecase/impl"
	"github.com/jnieto01/utils-01/database"
	utilsjwt "github.com/jnieto01/utils-01/jwt"
	"github.com/jnieto01/utils-01/logger"
	"github.com/jnieto01/utils-01/migration"
	"github.com/jnieto01/utils-01/rabbitmq"
	"github.com/redis/go-redis/v9"
)

const pathConfig = "./config/"

func main() {
	cfg := appconfig.Load(pathConfig)

	logger.SetupOut(cfg.Server.GoEnv)
	logger.SetService("payments-ms")
	if strings.ToUpper(cfg.Server.GoEnv) == "PROD" {
		gin.SetMode(gin.ReleaseMode)
	}
	logger.Info("Starting payments-ms-01...")

	mysqlDB, err := database.NewDatabase(cfg.MySQL)
	if err != nil {
		logger.Fatal("Failed to create database", err)
	}

	if err := migration.RunAutoMigrations(&cfg.MySQL, cfg.MySQLMigration); err != nil {
		logger.Fatal("Failed to run migrations", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
		PoolSize: cfg.RedisPoolSize,
	})

	jwtService := utilsjwt.NewJWTService(cfg.JWT.SecretKey)

	paymentRepo := repoImpl.NewPaymentRepository(mysqlDB.DB)
	trialRepo := repoImpl.NewTrialRepository(mysqlDB.DB)
	planRepo := repoImpl.NewPlanRepository(mysqlDB.DB)
	commissionRepo := repoImpl.NewMarketplaceCommissionRepository(mysqlDB.DB)
	paymentUC := impl.NewPaymentUsecase(paymentRepo, trialRepo, planRepo, commissionRepo, redisClient, cfg.MercadoPago, cfg.RabbitMQ)
	paymentHandler := handler.NewPaymentHandler(paymentUC)

	if cfg.RabbitMQ.Host != "" {
		go startManualAdPaymentConsumer(cfg.RabbitMQ, paymentUC)
	}

	r := router.SetupRouter(paymentHandler, jwtService)

	logger.Info("payments-ms-01 listening on port %d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		logger.Fatal("server error", err)
	}
}

func startManualAdPaymentConsumer(rmqCfg rabbitmq.Config, uc usecase.PaymentUsecase) {
	consumer, err := rabbitmq.NewMessageService(rmqCfg, "advertising.payment.manual.confirmed")
	if err != nil {
		logger.Warn("manual-ad-payment consumer: failed to connect to RabbitMQ", err)
		return
	}
	logger.Info("manual-ad-payment consumer: listening on queue advertising.payment.manual.confirmed")
	if err := consumer.Consume(context.Background(), "payments-ms-ad-manual-consumer", func(body []byte) error {
		var msg usecase.AdvertisingPaymentConfirmedMsg
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("invalid advertising.payment.manual.confirmed payload: %w", err)
		}
		return uc.HandleManualAdPaymentConfirmed(context.Background(), msg)
	}); err != nil {
		logger.Error("manual-ad-payment consumer failed", err)
	}
}
