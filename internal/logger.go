package internal

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func setupLogger() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	logLevel := logrus.InfoLevel
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "trace":
		logLevel = logrus.TraceLevel
	case "debug":
		logLevel = logrus.DebugLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	}

	Logger.SetLevel(logLevel)
}
