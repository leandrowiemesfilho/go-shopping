package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/api-gateway/internal/util"
	"github.com/leandrowiemesfilho/api-gateway/pkg/errors"
)

type ServiceProxy struct {
	target  *url.URL
	proxy   *httputil.ReverseProxy
	timeout time.Duration
	logger  *util.Logger
}

func NewServiceProxy(targetURL string, timeout time.Duration, logger *util.Logger) (*ServiceProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, errors.NewInternalError("Invalid target URL", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Customize the reverse proxy error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.LogError(err, map[string]interface{}{
			"url":    targetURL,
			"method": r.Method,
			"path":   r.URL.Path,
		})

		// Write error response
		statusCode, response := errors.ErrorResponse(
			errors.NewInternalError("Service unavailable", err),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if jsonResponse, marshalErr := json.Marshal(response); marshalErr == nil {
			w.Write(jsonResponse)
		}
	}

	// Modify response to log and handle errors
	proxy.ModifyResponse = func(resp *http.Response) error {
		logger.LogInfo("Service response", map[string]interface{}{
			"url":         targetURL,
			"method":      resp.Request.Method,
			"path":        resp.Request.URL.Path,
			"status":      resp.StatusCode,
			"status_text": resp.Status,
		})

		// Handle non-2xx responses
		if resp.StatusCode >= 400 {
			// Read the response body to log the error
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.NewInternalError("Failed to read error response", err)
			}

			// Log the error response from the service
			logger.LogError(fmt.Errorf("service returned error: %s", resp.Status), map[string]interface{}{
				"url":           targetURL,
				"method":        resp.Request.Method,
				"path":          resp.Request.URL.Path,
				"status_code":   resp.StatusCode,
				"response_body": string(body),
			})

			// Replace the body so it can be read again
			resp.Body = io.NopCloser(bytes.NewReader(body))
		}

		return nil
	}

	return &ServiceProxy{
		target:  target,
		proxy:   proxy,
		timeout: timeout,
		logger:  logger,
	}, nil
}

func (p *ServiceProxy) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), p.timeout)
		defer cancel()

		// Update request with timeout context
		c.Request = c.Request.WithContext(ctx)

		// Log the request
		p.logger.LogInfo("Proxying request", map[string]interface{}{
			"url":    p.target.String(),
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
		})

		// Serve the request
		p.proxy.ServeHTTP(c.Writer, c.Request)

		// Log the response status
		p.logger.LogInfo("Request completed", map[string]interface{}{
			"url":         p.target.String(),
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status_code": c.Writer.Status(),
			"client_ip":   c.ClientIP(),
		})
	}
}

// HealthCheck handler
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "api-gateway",
		"version":   "1.0.0",
	})
}

// NotFound handler for undefined routes
func NotFoundHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Endpoint not found",
		"code":  http.StatusNotFound,
		"path":  c.Request.URL.Path,
	})
}

// MethodNotAllowed handler
func MethodNotAllowedHandler(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error":  "Method not allowed",
		"code":   http.StatusMethodNotAllowed,
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})
}
