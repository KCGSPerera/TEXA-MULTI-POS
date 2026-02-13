package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

func RegisterInventoryRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h *handlers.InventoryHandler) {
	group := router.Group("/api/branches/:branch_id/inventory")
	group.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))

	group.GET("", h.List)
	group.GET("/:product_id", middleware.ValidateUUIDParams("product_id"), h.GetByProduct)
	group.POST("/adjust", authz.RequireRole("admin"), h.Adjust)
}
