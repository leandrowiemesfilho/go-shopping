package util

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leandrowiemesfilho/api-gateway/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	*zerolog.Logger
}

func NewLogger(cfg *config.LoggingConfig) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var logger zerolog.Logger

	// Set output format
	if cfg.Format == "text" {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	// Set output destination
	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to open log file, using stdout")
		} else {
			logger = logger.Output(file)
		}
	}

	return &Logger{&logger}
}

// GinLogger middleware for structured logging
func (l *Logger) GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log details
		end := time.Now()
		latency := end.Sub(start)
		if raw != "" {
			path = path + "?" + raw
		}

		logger := l.Info()
		if c.Writer.Status() >= 400 {
			logger = l.Error()
		}

		logger.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Dur("latency", latency).
			Str("request_id", c.GetString("request_id")).
			Msg("HTTP request")
	}
}

// LogError logs errors with context
func (l *Logger) LogError(err error, context map[string]interface{}) {
	logEvent := l.Error().Err(err)
	for key, value := range context {
		logEvent = logEvent.Interface(key, value)
	}
	logEvent.Msg("Application error")
}

// LogInfo logs info messages with context
func (l *Logger) LogInfo(message string, context map[string]interface{}) {
	logEvent := l.Info()
	for key, value := range context {
		logEvent = logEvent.Interface(key, value)
	}
	logEvent.Msg(message)
}
