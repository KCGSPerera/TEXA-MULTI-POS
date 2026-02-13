package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type RefundHandler struct{ service services.RefundService }

func NewRefundHandler(service services.RefundService) *RefundHandler {
	return &RefundHandler{service: service}
}

func (h *RefundHandler) Create(c *gin.Context) {
	var req models.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create refund")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}
