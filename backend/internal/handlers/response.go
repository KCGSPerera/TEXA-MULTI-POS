package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/utils"
)

type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func respondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, apiResponse{Success: true, Data: data})
}

func respondError(c *gin.Context, status int, errMsg string) {
	c.JSON(status, apiResponse{Success: false, Error: errMsg})
}

func respondNoContent(c *gin.Context) {
	c.JSON(http.StatusOK, apiResponse{Success: true})
}

func respondMappedError(c *gin.Context, err error, fallback string) {
	status, msg := utils.MapServiceError(err)
	if status == http.StatusInternalServerError {
		msg = fallback
	}
	respondError(c, status, msg)
}

func parsePagination(c *gin.Context) (int, int) {
	limit := models.DefaultLimit
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	if limit <= 0 {
		limit = models.DefaultLimit
	}
	if limit > models.MaxLimit {
		limit = models.MaxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}
