package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type InventoryHandler struct{ service services.InventoryService }

func NewInventoryHandler(service services.InventoryService) *InventoryHandler {
	return &InventoryHandler{service: service}
}

func (h *InventoryHandler) List(c *gin.Context) {
	items, err := h.service.ListByBranch(c.Request.Context(), c.Param("branch_id"))
	if err != nil {
		respondMappedError(c, err, "failed to list inventory")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *InventoryHandler) GetByProduct(c *gin.Context) {
	item, err := h.service.GetByProduct(c.Request.Context(), c.Param("branch_id"), c.Param("product_id"))
	if err != nil {
		respondMappedError(c, err, "failed to fetch inventory")
		return
	}
	respondSuccess(c, http.StatusOK, item)
}

func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req models.AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Adjust(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to adjust inventory")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}
