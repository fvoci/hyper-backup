// 📄 utiles/checker.go

package utilities

import (
	"os"
)

// CheckConfig는 백업 및 로그 회전에 필요한 주요 환경 변수들의 설정 상태를 점검하고, 각 서비스의 구성 여부를 로그로 기록합니다.
// rsync, rclone, MySQL, PostgreSQL, MongoDB, Traefik 로그 회전 설정을 확인하며, 필수 값이 누락된 경우 경고를 남깁니다.
// 백업 서비스가 하나도 구성되지 않았거나 원격 백업이 설정되지 않은 경우 추가 경고를 출력합니다.
// 항상 nil을 반환합니다.
func CheckConfig() error {
	Logger.Info("[HyperBackup] 🔍 Checking environment variables")

	var configured int
	var useRseries int

	// rsync
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	switch {
	case src != "" && dest != "":
		Logger.Info("[HyperBackup] ✅ Rsync backup configured")
		configured++
		useRseries++
	case src != "" || dest != "":
		Logger.Warn("[HyperBackup] ⚠️ Rsync: RSYNC_SRC or RSYNC_DEST is missing")
	}

	// rclone
	remote := os.Getenv("RCLONE_REMOTE")
	path := os.Getenv("RCLONE_PATH")
	switch {
	case remote != "" && path != "":
		Logger.Info("[HyperBackup] ✅ Rclone backup configured")
		configured++
		useRseries++
	case remote != "" || path != "":
		Logger.Warn("[HyperBackup] ⚠️ Rclone: RCLONE_REMOTE or RCLONE_PATH is missing")
	}

	// MySQL
	if os.Getenv("MYSQL_HOST") != "" {
		Logger.Info("[HyperBackup] ✅ MySQL backup configured")
		configured++
	}

	// PostgreSQL
	if os.Getenv("POSTGRES_HOST") != "" {
		Logger.Info("[HyperBackup] ✅ PostgreSQL backup configured")
		configured++
	}

	// MongoDB
	if os.Getenv("MONGO_HOST") != "" {
		Logger.Info("[HyperBackup] ✅ MongoDB backup configured")
		configured++
	}

	// Traefik
	if os.Getenv("TRAEFIK_LOG_FILE") != "" {
		Logger.Info("[HyperBackup] ✅ Traefik logrotate enabled")
		configured++
	}

	if useRseries == 0 {
		Logger.Warn("[HyperBackup] ⚠️ Warn: BACKUP WILL BE STORED LOCALLY ONLY")
	}

	if configured == 0 {
		Logger.Warn("[HyperBackup] 🤷 No backup services configured; nothing to do")
	}

	Logger.Info("[HyperBackup] ✅ Configuration check complete")
	LogDivider()
	return nil
}
