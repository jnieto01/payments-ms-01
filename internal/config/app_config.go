package config

import (
	"log"
	"sync"

	"github.com/jnieto01/utils-01/database"
	"github.com/jnieto01/utils-01/migration"
	"github.com/jnieto01/utils-01/rabbitmq"
	"github.com/spf13/viper"
)

const (
	configName = "application"
	configType = "yaml"
)

type MercadoPagoConfig struct {
	AccessToken     string
	WebhookSecret   string
	BaseURL         string
	NotificationURL string
	BackURL         string
}

type JWTConfig struct {
	SecretKey string
}

type ServerConfig struct {
	Host        string
	Port        int
	GoEnv       string
	ContextPath string
}

type Config struct {
	Server         ServerConfig
	MySQL          database.Config
	MySQLMigration migration.Config
	MercadoPago    MercadoPagoConfig
	JWT            JWTConfig
	RabbitMQ       rabbitmq.Config
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisPoolSize  int
}

var (
	cfg     *Config
	onceCfg sync.Once
)

func Load(path string) *Config {
	onceCfg.Do(func() {
		viper.SetConfigName(configName)
		viper.SetConfigType(configType)
		viper.AddConfigPath(path)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("fatal error reading config: %v", err)
		}

		cfg = &Config{}

		cfg.Server.Host = viper.GetString("server.host")
		cfg.Server.Port = viper.GetInt("server.port")
		cfg.Server.GoEnv = viper.GetString("server.goEnv")
		cfg.Server.ContextPath = viper.GetString("server.context-path")

		cfg.MySQL = database.Config{
			Type:            database.MySQL,
			Host:            viper.GetString("database.mysql.host"),
			Port:            viper.GetString("database.mysql.port"),
			User:            viper.GetString("database.mysql.user"),
			Password:        viper.GetString("database.mysql.password"),
			DBName:          viper.GetString("database.mysql.name"),
			MaxOpenConns:    viper.GetInt("database.mysql.maxOpenConns"),
			MaxIdleConns:    viper.GetInt("database.mysql.maxIdleConns"),
			ConnMaxLifetime: viper.GetString("database.mysql.connMaxLifetime"),
			ConnMaxIdleTime: viper.GetString("database.mysql.connMaxIdleTime"),
		}

		cfg.MySQLMigration = migration.Config{
			Enabled:   viper.GetBool("golang-migrate.mysql.enabled"),
			Locations: viper.GetString("golang-migrate.mysql.locations"),
		}

		cfg.MercadoPago = MercadoPagoConfig{
			AccessToken:     viper.GetString("mercadopago.access_token"),
			WebhookSecret:   viper.GetString("mercadopago.webhook_secret"),
			BaseURL:         viper.GetString("mercadopago.base_url"),
			NotificationURL: viper.GetString("mercadopago.notification_url"),
			BackURL:         viper.GetString("mercadopago.back_url"),
		}

		// JWT secret — must match the key used by user-ms to sign tokens.
		cfg.JWT = JWTConfig{
			SecretKey: viper.GetString("jwt.secret.key"),
		}

		cfg.RedisAddr = viper.GetString("data.redis.host") + ":" + viper.GetString("data.redis.port")
		cfg.RedisPassword = viper.GetString("data.redis.password")
		cfg.RedisDB = viper.GetInt("data.redis.db")
		cfg.RedisPoolSize = viper.GetInt("data.redis.pool-size")

		cfg.RabbitMQ = rabbitmq.Config{
			Host:     viper.GetString("rabbitmq.host"),
			Port:     viper.GetString("rabbitmq.port"),
			Username: viper.GetString("rabbitmq.username"),
			Password: viper.GetString("rabbitmq.password"),
			SSL:      rabbitmq.SSLConfig{Enabled: viper.GetBool("rabbitmq.ssl.enabled")},
		}
	})

	return cfg
}
