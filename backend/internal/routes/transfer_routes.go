package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

func RegisterTransferRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h *handlers.TransferHandler) {
	group := router.Group("/api/branches/:branch_id/transfers")
	group.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))
	group.POST("", authz.RequirePermission("transfer.create"), h.Create)
	group.POST("/:transfer_id/complete", middleware.ValidateUUIDParams("transfer_id"), authz.RequirePermission("transfer.complete"), h.Complete)
}
