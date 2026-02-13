package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type ProductHandler struct{ service services.ProductService }

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create product")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	out, err := h.service.GetByID(c.Request.Context(), c.Param("branch_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to fetch product")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *ProductHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	items, err := h.service.ListByBranch(c.Request.Context(), c.Param("branch_id"), limit, offset)
	if err != nil {
		respondMappedError(c, err, "failed to list products")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Update(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to update product")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to delete product")
		return
	}
	respondNoContent(c)
}
