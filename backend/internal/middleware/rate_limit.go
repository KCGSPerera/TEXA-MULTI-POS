package middleware

import (
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	visitors map[string]*visitor
	mu       sync.Mutex
	rps      rate.Limit
	burst    int
}

func RateLimitMiddleware() gin.HandlerFunc {
	rps := parseEnvFloat("RATE_LIMIT_RPS", 10)
	burst := parseEnvInt("RATE_LIMIT_BURST", 20)

	rl := &rateLimiter{visitors: make(map[string]*visitor), rps: rate.Limit(rps), burst: burst}
	go rl.cleanup(5 * time.Minute)

	skip := map[string]struct{}{
		"/health":  {},
		"/ready":   {},
		"/metrics": {},
	}

	return func(c *gin.Context) {
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		ip := c.ClientIP()
		if !rl.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"error":      "rate limit exceeded",
				"request_id": c.GetString("request_id"),
			})
			return
		}

		c.Next()
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, found := rl.visitors[ip]
	if !found {
		v = &visitor{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter.Allow()
}

func (rl *rateLimiter) cleanup(expiration time.Duration) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cutoff := time.Now().Add(-expiration)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func parseEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
