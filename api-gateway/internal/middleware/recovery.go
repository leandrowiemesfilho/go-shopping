package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/api-gateway/internal/util"
)

func RecoveryMiddleware(logger *util.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				stack := debug.Stack()
				logger.Error().
					Interface("error", err).
					Str("stack", string(stack)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("client_ip", c.ClientIP()).
					Str("user_agent", c.Request.UserAgent()).
					Msg("Panic recovered")

				// Check if the connection is still available
				if c.Writer.Status() == 0 {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Internal server error",
						"code":  http.StatusInternalServerError,
					})
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}
