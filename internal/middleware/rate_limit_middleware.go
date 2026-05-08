package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipLimiter struct {
	mu        sync.Mutex
	tokens    float64
	maxTokens float64
	rate      float64
	lastTime  time.Time
}

func newIPLimiter(ratePerSec, burst float64) *ipLimiter {
	return &ipLimiter{
		tokens:    burst,
		maxTokens: burst,
		rate:      ratePerSec,
		lastTime:  time.Now(),
	}
}

func (l *ipLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	elapsed := time.Since(l.lastTime).Seconds()
	l.lastTime = time.Now()
	l.tokens += elapsed * l.rate
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

var limitersMap sync.Map

func getIPLimiter(ip string) *ipLimiter {
	v, _ := limitersMap.LoadOrStore(ip, newIPLimiter(10, 30))
	return v.(*ipLimiter)
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !getIPLimiter(c.ClientIP()).allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
