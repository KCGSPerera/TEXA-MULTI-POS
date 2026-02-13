package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
)

func RegisterAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler, jwtMiddleware gin.HandlerFunc) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", jwtMiddleware, authHandler.Me)
	}
}
