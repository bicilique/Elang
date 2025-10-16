package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	logger := logrus.New()

	levelStr := os.Getenv("LOG_LEVEL") // e.g. "4" or "info"
	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		levelInt = int(logrus.InfoLevel) // default
	}

	logger.SetLevel(logrus.Level(levelInt))
	logger.SetFormatter(&logrus.JSONFormatter{})

	return logger
}
