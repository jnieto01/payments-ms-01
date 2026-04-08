module github.com/jnieto01/payments-ms

go 1.24.1

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/go-playground/validator/v10 v10.27.0
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/google/uuid v1.6.0
	github.com/jnieto01/utils-01 v1.0.32
	github.com/pressly/goose/v3 v3.24.3
	github.com/redis/go-redis/v9 v9.14.0
	github.com/spf13/viper v1.20.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/gorm v1.31.0
)

replace github.com/jnieto01/utils-01 => ../utils-01
