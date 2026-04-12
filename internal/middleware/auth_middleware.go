package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	utilsjwt "github.com/jnieto01/utils-01/jwt"
)

// AuthMiddleware validates the JWT issued by user-ms using the shared utils-01/jwt service.
// Token is read from the Authorization header (Bearer) or the "jwt" cookie.
func AuthMiddleware(jwtService utilsjwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ""

		// 1. Prefer Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Fall back to jwt cookie
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token: " + err.Error()})
			return
		}

		// Populate context from the "data" map in the JWT payload (same structure as user-ms).
		data := claims.GetData()
		if userID, ok := data["userId"].(string); ok {
			c.Set("user_id", userID)
		}
		if email, ok := data["email"].(string); ok {
			c.Set("user_email", email)
		}

		c.Next()
	}
}
