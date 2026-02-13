package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/cache"
)

type HealthHandler struct {
	db    *pgxpool.Pool
	cache cache.Client
}

func NewHealthHandler(db *pgxpool.Pool, cache cache.Client) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

func (h *HealthHandler) Health(c *gin.Context) {
	respondSuccess(c, http.StatusOK, gin.H{"status": "ok", "request_id": c.GetString("request_id")})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	checks := gin.H{"database": "ok", "redis": "disabled"}
	status := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = err.Error()
		status = http.StatusServiceUnavailable
	}

	if h.cache != nil && h.cache.Enabled() {
		if err := h.cache.Ping(ctx); err != nil {
			checks["redis"] = err.Error()
			status = http.StatusServiceUnavailable
		} else {
			checks["redis"] = "ok"
		}
	}

	if status != http.StatusOK {
		c.JSON(status, gin.H{"success": false, "error": "service not ready", "checks": checks, "request_id": c.GetString("request_id")})
		return
	}

	respondSuccess(c, http.StatusOK, gin.H{"status": "ready", "checks": checks, "request_id": c.GetString("request_id")})
}
