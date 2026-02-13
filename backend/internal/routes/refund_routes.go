package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

func RegisterRefundRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h *handlers.RefundHandler) {
	group := router.Group("/api/branches/:branch_id/refunds")
	group.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))
	group.POST("", authz.RequirePermission("refund.create"), h.Create)
}
