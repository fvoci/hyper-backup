package utilities

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance used throughout the application.
// It is pre-configured with text formatting, log level (default: info),
// and supports colored output and timestamped entries.
// Log level can be overridden using the LOG_LEVEL environment variable.
var Logger = newLogger()

// newLogger는 표준 출력으로 로그를 출력하고, 텍스트 포맷터와 환경 변수 LOG_LEVEL에 따라 로그 레벨을 설정한 새로운 logrus.Logger 인스턴스를 반환합니다.
// LOG_LEVEL이 유효하지 않으면 info 레벨로 기본 설정됩니다.
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
	levelStr := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
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

// LogDivider는 로그에 시각적인 구분선을 출력하여 로그 항목을 구분합니다.
func LogDivider() {
	Logger.Info("════════════════════════════════════════════════════════════════")
}
