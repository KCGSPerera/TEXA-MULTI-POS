package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type TransferHandler struct{ service services.TransferService }

func NewTransferHandler(service services.TransferService) *TransferHandler {
	return &TransferHandler{service: service}
}

func (h *TransferHandler) Create(c *gin.Context) {
	var req models.CreateBranchTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create transfer")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *TransferHandler) Complete(c *gin.Context) {
	out, err := h.service.Complete(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("transfer_id"))
	if err != nil {
		respondMappedError(c, err, "failed to complete transfer")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}
