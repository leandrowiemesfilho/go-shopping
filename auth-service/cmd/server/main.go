package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/auth-service/internal/config"
	"github.com/leandrowiemesfilho/auth-service/internal/database"
	"github.com/leandrowiemesfilho/auth-service/internal/handler"
	"github.com/leandrowiemesfilho/auth-service/internal/repository"
	"github.com/leandrowiemesfilho/auth-service/internal/service"
	"github.com/leandrowiemesfilho/auth-service/internal/util"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	if err := util.InitLogger(cfg.Logger.Level, cfg.Logger.Format); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		util.Error("Failed to connect to database", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		util.Error("Failed to run migrations", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Migrations failed: %v", err)
	}

	// Initialize utilities
	jwtUtil := util.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.Issuer)
	passwordUtil := util.NewPasswordUtil()

	// Initialize repository
	userRepo := repository.NewUserRepository(db.Pool)

	// Initialize service
	authService := service.NewAuthService(
		userRepo,
		jwtUtil,
		passwordUtil,
		&service.JWTConfig{
			Secret:          cfg.JWT.Secret,
			ExpirationHours: cfg.JWT.ExpirationHours,
			Issuer:          cfg.JWT.Issuer,
		},
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)

	// Setup router
	router := setupRouter(authHandler)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout * time.Second,
		WriteTimeout: cfg.Server.WriteTimeout * time.Second,
	}

	// Start server in a goroutine
	go func() {
		util.Info("Starting auth service", map[string]interface{}{
			"port": cfg.Server.Port,
			"mode": cfg.Server.Mode,
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			util.Error("Failed to start server", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	util.Info("Shutting down server...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		util.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Server shutdown failed: %v", err)
	}

	util.Info("Server exited properly", nil)
}

func setupRouter(authHandler *handler.AuthHandler) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Routes
	router.GET("/health", authHandler.HealthCheck)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	return router
}
