package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/product-service/internal/config"
	"github.com/leandrowiemesfilho/product-service/internal/database"
	"github.com/leandrowiemesfilho/product-service/internal/handler"
	"github.com/leandrowiemesfilho/product-service/internal/repository"
	"github.com/leandrowiemesfilho/product-service/internal/service"
	"github.com/leandrowiemesfilho/product-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	db, err := database.NewDB(&cfg.Database, appLogger.SugaredLogger)
	if err != nil {
		appLogger.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.InitSchema(appLogger.SugaredLogger); err != nil {
		appLogger.Fatalw("Failed to initialize database schema", "error", err)
	}

	// Initialize repository, service, and handlers
	productRepo := repository.NewProductRepository(db.DB, appLogger.SugaredLogger)
	productService := service.NewProductService(productRepo, appLogger.SugaredLogger)
	productHandler := handler.NewProductHandler(productService, appLogger.SugaredLogger)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())

	// Routes
	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		products := api.Group("/products")
		{
			products.GET("", productHandler.GetAllProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}
	}

	appLogger.Infow("Starting server", "port", cfg.Server.Port)
	if err := router.Run(cfg.Server.Port); err != nil {
		appLogger.Fatalw("Failed to start server", "error", err)
	}
}
