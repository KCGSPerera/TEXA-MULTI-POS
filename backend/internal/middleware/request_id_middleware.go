package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/requestctx"
)

const maxRequestIDLen = 64

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader("X-Request-Id"))
		if requestID == "" {
			requestID = uuid.NewString()
		}
		if len(requestID) > maxRequestIDLen {
			requestID = requestID[:maxRequestIDLen]
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-Id", requestID)
		c.Request = c.Request.WithContext(requestctx.WithRequestID(c.Request.Context(), requestID))
		c.Next()
	}
}
