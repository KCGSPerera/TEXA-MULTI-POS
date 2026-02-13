package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

func RegisterReportRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h *handlers.ReportHandler) {
	group := router.Group("/api/branches/:branch_id/reports")
	group.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))

	group.GET("/daily-sales-summary", h.DailySalesSummary)
	group.GET("/product-sales", h.ProductSalesReport)
	group.GET("/payment-method-summary", h.PaymentMethodSummary)
	group.GET("/stock-valuation", h.StockValuationReport)
}
