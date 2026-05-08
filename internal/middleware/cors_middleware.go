package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var defaultAllowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:3001",
	"http://localhost:3002",
	"http://localhost:3003",
	"https://nvfsports.com",
	"https://www.nvfsports.com",
	"https://nvfsports.local",
	"https://www.nvfsports.local",
	"https://admin.nvfsports.com",
	"https://admin.nvfsports.local",
	"https://app.nvfsports.com",
	"https://app.nvfsports.local",
}

func buildAllowedOrigins() map[string]struct{} {
	set := make(map[string]struct{})
	if env := os.Getenv("CORS_ALLOWED_ORIGINS"); env != "" {
		for _, o := range strings.Split(env, ",") {
			if o = strings.TrimSpace(o); o != "" {
				set[o] = struct{}{}
			}
		}
	} else {
		for _, o := range defaultAllowedOrigins {
			set[o] = struct{}{}
		}
	}
	return set
}

func CORSMiddleware() gin.HandlerFunc {
	allowed := buildAllowedOrigins()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Vary", "Origin")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Signature")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
