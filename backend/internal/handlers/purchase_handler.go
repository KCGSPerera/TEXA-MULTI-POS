package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type PurchaseHandler struct{ service services.PurchaseService }

func NewPurchaseHandler(service services.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{service: service}
}

func (h *PurchaseHandler) CreatePurchaseOrder(c *gin.Context) {
	var req models.CreatePurchaseOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.CreatePurchaseOrder(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create purchase order")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *PurchaseHandler) ListPurchaseOrders(c *gin.Context) {
	limit, offset := parsePagination(c)
	items, err := h.service.ListPurchaseOrders(c.Request.Context(), c.Param("branch_id"), limit, offset)
	if err != nil {
		respondMappedError(c, err, "failed to list purchase orders")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *PurchaseHandler) CreateGRN(c *gin.Context) {
	var req models.CreateGRNRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.CreateGRN(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("purchase_order_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create GRN")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *PurchaseHandler) ApproveGRN(c *gin.Context) {
	out, err := h.service.ApproveGRN(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("grn_id"))
	if err != nil {
		respondMappedError(c, err, "failed to approve GRN")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}
