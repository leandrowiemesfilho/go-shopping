package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/api-gateway/internal/config"
	"github.com/leandrowiemesfilho/api-gateway/internal/handler"
	"github.com/leandrowiemesfilho/api-gateway/internal/middleware"
	"github.com/leandrowiemesfilho/api-gateway/internal/util"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := util.NewLogger(&config.AppConfig.Logging)
	defer logger.Info().Msg("Shutting down API Gateway")

	// Set Gin mode
	gin.SetMode(config.AppConfig.Server.Mode)

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggingMiddleware(logger))

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppConfig.Cors.AllowedOrigins,
		AllowMethods:     config.AppConfig.Cors.AllowedMethods,
		AllowHeaders:     config.AppConfig.Cors.AllowedHeaders,
		AllowCredentials: config.AppConfig.Cors.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", handler.HealthCheck)

	// Service proxies
	authProxy, err := handler.NewServiceProxy(
		config.AppConfig.Services.Auth.BaseURL,
		config.AppConfig.Services.Auth.Timeout*time.Second,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create auth service proxy")
	}

	productsProxy, err := handler.NewServiceProxy(
		config.AppConfig.Services.Products.BaseURL,
		config.AppConfig.Services.Products.Timeout*time.Second,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create products service proxy")
	}

	// Routes
	api := router.Group("/api/v1")
	{
		// Auth routes (no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authProxy.Handler())
			auth.POST("/login", authProxy.Handler())
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(config.AppConfig.Auth.JWTSecret, logger))
		{
			// Product routes
			products := protected.Group("/products")
			{
				products.GET("", productsProxy.Handler())
				products.GET("/:id", productsProxy.Handler())
				products.POST("", productsProxy.Handler())
				products.PUT("/:id", productsProxy.Handler())
				products.DELETE("/:id", productsProxy.Handler())
			}
		}
	}

	// Custom handlers for 404 and 405
	router.NoRoute(handler.NotFoundHandler)
	router.NoMethod(handler.MethodNotAllowedHandler)

	// Create server with timeouts
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler:      router,
		ReadTimeout:  config.AppConfig.Server.ReadTimeout * time.Second,
		WriteTimeout: config.AppConfig.Server.WriteTimeout * time.Second,
		IdleTimeout:  config.AppConfig.Server.IdleTimeout * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info().Str("port", strconv.Itoa(config.AppConfig.Server.Port)).Msg("Starting API Gateway")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited properly")
}
