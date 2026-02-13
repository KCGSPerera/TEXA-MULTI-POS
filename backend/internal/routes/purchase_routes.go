package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

func RegisterPurchaseRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h *handlers.PurchaseHandler) {
	group := router.Group("/api/branches/:branch_id")
	group.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))

	group.GET("/purchase-orders", h.ListPurchaseOrders)
	group.POST("/purchase-orders", h.CreatePurchaseOrder)
	group.POST("/purchase-orders/:purchase_order_id/grns", middleware.ValidateUUIDParams("purchase_order_id"), h.CreateGRN)
	group.POST("/grns/:grn_id/approve", middleware.ValidateUUIDParams("grn_id"), authz.RequireRole("admin"), h.ApproveGRN)
}
