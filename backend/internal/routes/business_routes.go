package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
)

type BusinessHandlers struct {
	CategoryHandler *handlers.CategoryHandler
	ProductHandler  *handlers.ProductHandler
	SupplierHandler *handlers.SupplierHandler
	CustomerHandler *handlers.CustomerHandler
	SaleHandler     *handlers.SaleHandler
}

func RegisterBusinessRoutes(router *gin.Engine, authz *middleware.AuthorizationMiddleware, jwtMiddleware gin.HandlerFunc, h BusinessHandlers) {
	branchGroup := router.Group("/api/branches/:branch_id")
	branchGroup.Use(jwtMiddleware, authz.RequireBranchMatch("branch_id"), middleware.ValidateUUIDParams("branch_id"))

	branchGroup.GET("/categories", h.CategoryHandler.List)
	branchGroup.GET("/categories/:id", middleware.ValidateUUIDParams("id"), h.CategoryHandler.GetByID)
	branchGroup.POST("/categories", h.CategoryHandler.Create)
	branchGroup.PUT("/categories/:id", middleware.ValidateUUIDParams("id"), h.CategoryHandler.Update)
	branchGroup.DELETE("/categories/:id", middleware.ValidateUUIDParams("id"), authz.RequireRole("admin"), h.CategoryHandler.Delete)

	branchGroup.GET("/products", h.ProductHandler.List)
	branchGroup.GET("/products/:id", middleware.ValidateUUIDParams("id"), h.ProductHandler.GetByID)
	branchGroup.POST("/products", h.ProductHandler.Create)
	branchGroup.PUT("/products/:id", middleware.ValidateUUIDParams("id"), h.ProductHandler.Update)
	branchGroup.DELETE("/products/:id", middleware.ValidateUUIDParams("id"), authz.RequireRole("admin"), h.ProductHandler.Delete)

	branchGroup.GET("/suppliers", h.SupplierHandler.List)
	branchGroup.GET("/suppliers/:id", middleware.ValidateUUIDParams("id"), h.SupplierHandler.GetByID)
	branchGroup.POST("/suppliers", h.SupplierHandler.Create)
	branchGroup.PUT("/suppliers/:id", middleware.ValidateUUIDParams("id"), h.SupplierHandler.Update)
	branchGroup.DELETE("/suppliers/:id", middleware.ValidateUUIDParams("id"), authz.RequireRole("admin"), h.SupplierHandler.Delete)

	branchGroup.GET("/customers", h.CustomerHandler.List)
	branchGroup.GET("/customers/:id", middleware.ValidateUUIDParams("id"), h.CustomerHandler.GetByID)
	branchGroup.POST("/customers", h.CustomerHandler.Create)
	branchGroup.PUT("/customers/:id", middleware.ValidateUUIDParams("id"), h.CustomerHandler.Update)
	branchGroup.DELETE("/customers/:id", middleware.ValidateUUIDParams("id"), authz.RequireRole("admin"), h.CustomerHandler.Delete)

	branchGroup.POST("/sales", authz.RequirePermission("sale.create"), h.SaleHandler.Create)
}
