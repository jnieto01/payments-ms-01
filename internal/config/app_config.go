package config

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type AppConfig struct {
	DB          *gorm.DB
	Redis       *redis.Client
	MercadoPago MercadoPagoConfig
}

type MercadoPagoConfig struct {
	AccessToken     string
	WebhookSecret   string
	BaseURL         string
	NotificationURL string
}

func LoadConfig() {
	viper.SetConfigName("application")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error reading config: %w", err))
	}
}

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("database.mysql.user"),
		viper.GetString("database.mysql.password"),
		viper.GetString("database.mysql.host"),
		viper.GetInt("database.mysql.port"),
		viper.GetString("database.mysql.name"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(viper.GetInt("database.mysql.maxOpenConns"))
	sqlDB.SetMaxIdleConns(viper.GetInt("database.mysql.maxIdleConns"))
	sqlDB.SetConnMaxLifetime(viper.GetDuration("database.mysql.connMaxLifetime") * time.Hour)
	sqlDB.SetConnMaxIdleTime(viper.GetDuration("database.mysql.connMaxIdleTime") * time.Minute)

	return db
}

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", viper.GetString("data.redis.host"), viper.GetString("data.redis.port")),
		Password: viper.GetString("data.redis.password"),
		DB:       viper.GetInt("data.redis.db"),
		PoolSize: viper.GetInt("data.redis.pool-size"),
	})
}

func LoadMercadoPagoConfig() MercadoPagoConfig {
	return MercadoPagoConfig{
		AccessToken:     viper.GetString("mercadopago.access_token"),
		WebhookSecret:   viper.GetString("mercadopago.webhook_secret"),
		BaseURL:         viper.GetString("mercadopago.base_url"),
		NotificationURL: viper.GetString("mercadopago.notification_url"),
	}
}
