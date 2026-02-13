package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/observability"
)

func MetricsMiddleware() gin.HandlerFunc {
	observability.RegisterMetrics()

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		observability.ObserveHTTP(c.Request.Method, path, strconv.Itoa(c.Writer.Status()), time.Since(start).Seconds())
	}
}
