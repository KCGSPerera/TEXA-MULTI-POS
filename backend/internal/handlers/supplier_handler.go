package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type SupplierHandler struct{ service services.SupplierService }

func NewSupplierHandler(service services.SupplierService) *SupplierHandler {
	return &SupplierHandler{service: service}
}

func (h *SupplierHandler) Create(c *gin.Context) {
	var req models.CreateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create supplier")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *SupplierHandler) GetByID(c *gin.Context) {
	out, err := h.service.GetByID(c.Request.Context(), c.Param("branch_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to fetch supplier")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *SupplierHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	items, err := h.service.ListByBranch(c.Request.Context(), c.Param("branch_id"), limit, offset)
	if err != nil {
		respondMappedError(c, err, "failed to list suppliers")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	var req models.UpdateSupplierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Update(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to update supplier")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to delete supplier")
		return
	}
	respondNoContent(c)
}
