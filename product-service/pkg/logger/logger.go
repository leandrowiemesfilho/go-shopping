package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level, format string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var encoderConfig zapcore.EncoderConfig
	if format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	var encoder zapcore.Encoder
	if format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugaredLogger := logger.Sugar()

	return &Logger{sugaredLogger}, nil
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var zapFields []interface{}
	for key, value := range fields {
		zapFields = append(zapFields, key, value)
	}
	return &Logger{l.SugaredLogger.With(zapFields...)}
}

// Helper methods for different log levels with structured logging
func (l *Logger) InfoWithFields(msg string, fields map[string]interface{}) {
	l.WithFields(fields).Info(msg)
}

func (l *Logger) ErrorWithFields(msg string, fields map[string]interface{}) {
	l.WithFields(fields).Error(msg)
}

func (l *Logger) WarnWithFields(msg string, fields map[string]interface{}) {
	l.WithFields(fields).Warn(msg)
}
