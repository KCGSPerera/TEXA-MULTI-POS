package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/config"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/routes"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

func main() {
	config.LoadEnv()
	logger.Init()

	databaseURL := os.Getenv("DATABASE_URL")
	database.Connect(databaseURL)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	userRepo := repositories.NewUserRepository(database.DB)
	roleRepo := repositories.NewRoleRepository(database.DB)
	auditRepo := repositories.NewAuditRepository(database.DB)
	auditService := services.NewAuditService(auditRepo)

	authService, err := services.NewAuthService(userRepo, jwtSecret)
	if err != nil {
		log.Fatal(err)
	}
	authHandler := handlers.NewAuthHandler(authService)
	jwtMiddleware := middleware.JWTAuthMiddleware(jwtSecret)
	authz := middleware.NewAuthorizationMiddleware(roleRepo)

	categoryRepo := repositories.NewCategoryRepository(database.DB)
	productRepo := repositories.NewProductRepository(database.DB)
	supplierRepo := repositories.NewSupplierRepository(database.DB)
	customerRepo := repositories.NewCustomerRepository(database.DB)
	saleRepo := repositories.NewSaleRepository(database.DB)
	inventoryRepo := repositories.NewInventoryRepository(database.DB)
	purchaseRepo := repositories.NewPurchaseRepository(database.DB)
	reportRepo := repositories.NewReportRepository(database.DB)

	categoryService := services.NewCategoryService(categoryRepo, auditService)
	productService := services.NewProductService(productRepo, categoryRepo, auditService)
	supplierService := services.NewSupplierService(supplierRepo, auditService)
	customerService := services.NewCustomerService(customerRepo, auditService)
	saleService := services.NewSaleService(database.DB, saleRepo, productRepo, inventoryRepo, auditService)
	inventoryService := services.NewInventoryService(database.DB, inventoryRepo, productRepo, auditService)
	purchaseService := services.NewPurchaseService(database.DB, purchaseRepo, supplierRepo, productRepo, inventoryRepo, auditService)
	reportService := services.NewReportService(reportRepo)

	businessHandlers := routes.BusinessHandlers{
		CategoryHandler: handlers.NewCategoryHandler(categoryService),
		ProductHandler:  handlers.NewProductHandler(productService),
		SupplierHandler: handlers.NewSupplierHandler(supplierService),
		CustomerHandler: handlers.NewCustomerHandler(customerService),
		SaleHandler:     handlers.NewSaleHandler(saleService),
	}
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	purchaseHandler := handlers.NewPurchaseHandler(purchaseService)
	reportHandler := handlers.NewReportHandler(reportService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.BodySizeLimitMiddleware(1 << 20))
	r.Use(middleware.RequestLoggerMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "TEXA MULTI POS API Running"})
	})

	routes.RegisterAuthRoutes(r, authHandler, jwtMiddleware)
	routes.RegisterBusinessRoutes(r, authz, jwtMiddleware, businessHandlers)
	routes.RegisterInventoryRoutes(r, authz, jwtMiddleware, inventoryHandler)
	routes.RegisterPurchaseRoutes(r, authz, jwtMiddleware, purchaseHandler)
	routes.RegisterReportRoutes(r, authz, jwtMiddleware, reportHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)
}
