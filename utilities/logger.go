package utilities

import (
	"os"

	"github.com/sirupsen/logrus"
)

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

	log.SetLevel(logrus.InfoLevel)

	return log
}

// LogDivider prints a visual divider line in the logs to separate log entries
func LogDivider() {
	Logger.Info("════════════════════════════════════════════════════════════════")
}
