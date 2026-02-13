package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/cache"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/config"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/database"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/handlers"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/jobs"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/logger"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/middleware"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/migration"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/observability"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/repositories"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/routes"
	"github.com/kcgsperera/texa-multi-pos/backend/internal/services"
)

func main() {
	config.LoadEnv()
	logger.Init()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	database.Connect(databaseURL)

	version, changed, err := migration.Run(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	logger.L().Info().Uint("migration_version", version).Bool("changed", changed).Msg("migrations_ready")

	redisClient, err := cache.NewRedisClient(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	traceShutdown, err := observability.InitTracing(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = traceShutdown(ctx)
	}()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	userRepo := repositories.NewUserRepository(database.DB)
	roleRepo := repositories.NewRoleRepository(database.DB, redisClient)
	auditRepo := repositories.NewAuditRepository(database.DB)
	ledgerRepo := repositories.NewLedgerRepository(database.DB)
	discountRepo := repositories.NewDiscountRepository(database.DB, redisClient)
	loyaltyRepo := repositories.NewLoyaltyRepository(database.DB)
	refundRepo := repositories.NewRefundRepository(database.DB)
	transferRepo := repositories.NewTransferRepository(database.DB)
	tokenRepo := repositories.NewTokenRepository(database.DB)
	accountRepo := repositories.NewAccountRepository(database.DB)

	auditService := services.NewAuditService(auditRepo)
	ledgerService := services.NewLedgerService(ledgerRepo, auditService)
	loyaltyService := services.NewLoyaltyService(loyaltyRepo)
	tokenService := services.NewTokenService(database.DB, tokenRepo)
	accountBootstrapService := services.NewAccountBootstrapService(database.DB, accountRepo)

	authService, err := services.NewAuthService(userRepo, tokenService, jwtSecret)
	if err != nil {
		log.Fatal(err)
	}
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler(database.DB, redisClient)
	jwtMiddleware := middleware.JWTAuthMiddleware(jwtSecret)
	authz := middleware.NewAuthorizationMiddleware(roleRepo)

	if err := accountBootstrapService.EnsureAllBranchesSeeded(context.Background()); err != nil {
		log.Fatal(err)
	}

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
	saleService := services.NewSaleService(database.DB, saleRepo, productRepo, inventoryRepo, discountRepo, loyaltyService, ledgerService, auditService)
	inventoryService := services.NewInventoryService(database.DB, inventoryRepo, productRepo, auditService)
	purchaseService := services.NewPurchaseService(database.DB, purchaseRepo, supplierRepo, productRepo, inventoryRepo, ledgerService, auditService)
	reportService := services.NewReportService(reportRepo)
	refundService := services.NewRefundService(database.DB, refundRepo, saleRepo, inventoryRepo, loyaltyService, ledgerService, auditService)
	transferService := services.NewTransferService(database.DB, transferRepo, inventoryRepo, productRepo, ledgerService, auditService)

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
	refundHandler := handlers.NewRefundHandler(refundService)
	transferHandler := handlers.NewTransferHandler(transferService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.BodySizeLimitMiddleware(1 << 20))
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.MetricsMiddleware())
	r.Use(observability.GinTracingMiddleware())

	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	routes.RegisterAuthRoutes(r, authHandler, jwtMiddleware)
	routes.RegisterBusinessRoutes(r, authz, jwtMiddleware, businessHandlers)
	routes.RegisterInventoryRoutes(r, authz, jwtMiddleware, inventoryHandler)
	routes.RegisterPurchaseRoutes(r, authz, jwtMiddleware, purchaseHandler)
	routes.RegisterReportRoutes(r, authz, jwtMiddleware, reportHandler)
	routes.RegisterRefundRoutes(r, authz, jwtMiddleware, refundHandler)
	routes.RegisterTransferRoutes(r, authz, jwtMiddleware, transferHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker := jobs.NewWorker(database.DB, redisClient, databaseURL)
	worker.Start(ctx)

	go func() {
		logger.L().Info().Str("port", port).Msg("server_start")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
