package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type CategoryHandler struct {
	service services.CategoryService
}

func NewCategoryHandler(service services.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create category")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
	out, err := h.service.GetByID(c.Request.Context(), c.Param("branch_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to fetch category")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *CategoryHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	items, err := h.service.ListByBranch(c.Request.Context(), c.Param("branch_id"), limit, offset)
	if err != nil {
		respondMappedError(c, err, "failed to list categories")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Update(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to update category")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to delete category")
		return
	}
	respondNoContent(c)
}
