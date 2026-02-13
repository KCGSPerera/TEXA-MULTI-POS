package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ValidateUUIDParams(names ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, name := range names {
			if _, err := uuid.Parse(c.Param(name)); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid UUID for parameter: " + name})
				return
			}
		}
		c.Next()
	}
}
