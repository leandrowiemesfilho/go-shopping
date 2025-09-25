package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/auth-service/internal/model"
	"github.com/leandrowiemesfilho/auth-service/internal/service"
	"github.com/leandrowiemesfilho/auth-service/internal/util"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error("Invalid registration request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid request payload",
		})
		return
	}

	authResponse, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		util.Error("Registration failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		statusCode := http.StatusInternalServerError
		errorMsg := "Registration failed"

		// You can add more specific error handling here
		if err.Error() == "email already registered" {
			statusCode = http.StatusConflict
			errorMsg = "Email already registered"
		}

		c.JSON(statusCode, model.ErrorResponse{
			Error: errorMsg,
		})
		return
	}

	util.Info("User registered successfully", map[string]interface{}{
		"email":   req.Email,
		"user_id": authResponse.User.ID,
	})

	c.JSON(http.StatusCreated, authResponse)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error("Invalid login request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "Invalid request payload",
		})
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		util.Warn("Login failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		})

		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error: "Invalid credentials",
		})
		return
	}

	util.Info("User logged in successfully", map[string]interface{}{
		"email":   req.Email,
		"user_id": authResponse.User.ID,
	})

	c.JSON(http.StatusOK, authResponse)
}

func (h *AuthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}
