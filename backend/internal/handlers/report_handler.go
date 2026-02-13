package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

type ReportHandler struct{ service services.ReportService }

func NewReportHandler(service services.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) DailySalesSummary(c *gin.Context) {
	from, to, ok := parseRange(c)
	if !ok {
		return
	}
	out, err := h.service.DailySalesSummary(c.Request.Context(), c.Param("branch_id"), from, to)
	if err != nil {
		respondMappedError(c, err, "failed to generate daily sales summary")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *ReportHandler) ProductSalesReport(c *gin.Context) {
	from, to, ok := parseRange(c)
	if !ok {
		return
	}
	out, err := h.service.ProductSalesReport(c.Request.Context(), c.Param("branch_id"), from, to)
	if err != nil {
		respondMappedError(c, err, "failed to generate product sales report")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *ReportHandler) PaymentMethodSummary(c *gin.Context) {
	from, to, ok := parseRange(c)
	if !ok {
		return
	}
	out, err := h.service.PaymentMethodSummary(c.Request.Context(), c.Param("branch_id"), from, to)
	if err != nil {
		respondMappedError(c, err, "failed to generate payment method summary")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func (h *ReportHandler) StockValuationReport(c *gin.Context) {
	out, err := h.service.StockValuationReport(c.Request.Context(), c.Param("branch_id"))
	if err != nil {
		respondMappedError(c, err, "failed to generate stock valuation report")
		return
	}
	respondSuccess(c, http.StatusOK, out)
}

func parseRange(c *gin.Context) (time.Time, time.Time, bool) {
	var q models.DateRangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return time.Time{}, time.Time{}, false
	}
	fromDate, err := time.Parse("2006-01-02", q.From)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid from date")
		return time.Time{}, time.Time{}, false
	}
	toDate, err := time.Parse("2006-01-02", q.To)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid to date")
		return time.Time{}, time.Time{}, false
	}
	if toDate.Before(fromDate) {
		respondError(c, http.StatusBadRequest, "to date must be greater than or equal to from date")
		return time.Time{}, time.Time{}, false
	}
	return fromDate.UTC(), toDate.Add(24 * time.Hour).UTC(), true
}
