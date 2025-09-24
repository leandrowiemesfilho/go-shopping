package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/api-gateway/internal/util"
	"github.com/leandrowiemesfilho/api-gateway/pkg/errors"
)

func AuthMiddleware(jwtSecret string, logger *util.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token (simplified - in real implementation, validate with auth service)
		userID, err := validateToken(tokenString, jwtSecret)
		if err != nil {
			logger.Warn().Str("path", c.Request.URL.Path).Msg("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Set user ID in context for downstream services
		c.Set("user_id", userID)
		c.Next()
	}
}

func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/health",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

func validateToken(tokenString, jwtSecret string) (string, error) {
	// Simplified token validation
	// In real implementation, this would:
	// 1. Validate JWT signature
	// 2. Check expiration
	// 3. Possibly call auth service for validation

	if tokenString == "" {
		return "", errors.NewUnauthorizedError("Empty token")
	}

	// Mock validation - replace with real JWT validation
	if len(tokenString) < 10 {
		return "", errors.NewUnauthorizedError("Invalid token format")
	}

	// Mock user ID extraction
	return "user-" + tokenString[:8], nil
}
