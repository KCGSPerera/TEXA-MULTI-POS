package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type SaleHandler struct{ service services.SaleService }

func NewSaleHandler(service services.SaleService) *SaleHandler { return &SaleHandler{service: service} }

func (h *SaleHandler) Create(c *gin.Context) {
	var req models.CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), c.Param("branch_id"), c.GetString("user_id"), req)
	if err != nil {
		respondMappedError(c, err, "failed to create sale")
		return
	}
	respondSuccess(c, http.StatusCreated, out)
}
