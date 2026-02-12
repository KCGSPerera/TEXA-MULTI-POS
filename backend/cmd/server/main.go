package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/config"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
)

func main() {
	config.LoadEnv()

	databaseURL := os.Getenv("DATABASE_URL")
	database.Connect(databaseURL)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "TEXA MULTI POS API Running",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("🚀 Server running on port", port)
	r.Run(":" + port)
}
