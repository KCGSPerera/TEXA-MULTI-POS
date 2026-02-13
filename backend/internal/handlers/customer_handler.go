package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type CustomerHandler struct{ service services.CustomerService }

func NewCustomerHandler(service services.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create customer")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	out, err := h.service.GetByID(c.Request.Context(), c.Param("branch_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to fetch customer")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *CustomerHandler) List(c *gin.Context) {
	limit, offset := parsePagination(c)
	items, err := h.service.ListByBranch(c.Request.Context(), c.Param("branch_id"), limit, offset)
	if err != nil {
		respondMappedError(c, err, "failed to list customers")
		return
	}
	respondSuccess(c, http.StatusOK, items)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Update(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to update customer")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), c.Param("id"))
	if err != nil {
		respondMappedError(c, err, "failed to delete customer")
		return
	}
	respondNoContent(c)
}
