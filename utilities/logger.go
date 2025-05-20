package utilities

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance used throughout the application.
// It is pre-configured with text formatting, log level (default: info),
// and supports colored output and timestamped entries.
// Log level can be overridden using the LOG_LEVEL environment variable.
var Logger = newLogger()

func newLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		PadLevelText:    true,
	})

	// Set log level from LOG_LEVEL env (default: info)
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		log.SetLevel(logrus.InfoLevel)
	} else {
		level, err := logrus.ParseLevel(levelStr)
		if err != nil {
			log.Warnf("⚠️ Invalid LOG_LEVEL: %s. Falling back to info", levelStr)
			log.SetLevel(logrus.InfoLevel)
		} else {
			log.SetLevel(level)
		}
	}

	return log
}

// LogDivider prints a visual divider line in the logs to separate log entries
func LogDivider() {
	Logger.Info("════════════════════════════════════════════════════════════════")
}
