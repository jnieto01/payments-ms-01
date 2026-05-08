package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	utilsjwt "github.com/jnieto01/utils-01/jwt"
)

const adminRole = "admin_nvf"

// AdminMiddleware validates the JWT and enforces the admin_nvf role.
func AdminMiddleware(jwtService utilsjwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if tokenStr == "" {
			if cookie, err := c.Cookie("jwt"); err == nil {
				tokenStr = cookie
			}
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := jwtService.ValidateToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		role, _ := claims.GetString("role")
		if role != adminRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: admin_nvf role required"})
			return
		}

		if userID, ok := claims.GetString("userId"); ok {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}
