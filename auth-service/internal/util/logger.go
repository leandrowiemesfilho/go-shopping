package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func InitLogger(level, format string) error {
	Logger = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	Logger.SetLevel(logLevel)

	// Set output format
	if format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	Logger.SetOutput(os.Stdout)
	return nil
}

func Info(msg string, fields map[string]interface{}) {
	Logger.WithFields(fields).Info(msg)
}

func Error(msg string, fields map[string]interface{}) {
	Logger.WithFields(fields).Error(msg)
}

func Warn(msg string, fields map[string]interface{}) {
	Logger.WithFields(fields).Warn(msg)
}

func Debug(msg string, fields map[string]interface{}) {
	Logger.WithFields(fields).Debug(msg)
}
