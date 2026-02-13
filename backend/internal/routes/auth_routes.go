package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
)

func RegisterAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, jwtMiddleware gin.HandlerFunc) {
	registerAuthGroup(router.Group("/auth"), authHandler, jwtMiddleware)
	registerAuthGroup(router.Group("/api/auth"), authHandler, jwtMiddleware)
}

func registerAuthGroup(group *gin.RouterGroup, authHandler *handlers.AuthHandler, jwtMiddleware gin.HandlerFunc) {
	group.POST("/register", authHandler.Register)
	group.POST("/login", authHandler.Login)
	group.POST("/refresh", authHandler.Refresh)
	group.POST("/logout", authHandler.Logout)
	group.GET("/me", jwtMiddleware, authHandler.Me)
}
