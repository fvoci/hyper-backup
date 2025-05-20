// 📄 utiles/checker.go

package utiles

import (
	"log"
	"os"
)

func CheckConfig() error {
	log.Printf("[HyperBackup] 🔍 Checking environment variables")

	var configured int
	var useRseries int

	// rsync
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	switch {
	case src != "" && dest != "":
		log.Printf("[HyperBackup] ✅ Rsync backup configured")
		configured++
		useRseries++
	case src != "" || dest != "":
		log.Printf("[HyperBackup] ⚠️ Rsync: RSYNC_SRC or RSYNC_DEST is missing")
	}

	// rclone
	remote := os.Getenv("RCLONE_REMOTE")
	path := os.Getenv("RCLONE_PATH")
	switch {
	case remote != "" && path != "":
		log.Printf("[HyperBackup] ✅ Rclone backup configured")
		configured++
		useRseries++
	case remote != "" || path != "":
		log.Printf("[HyperBackup] ⚠️ Rclone: RCLONE_REMOTE or RCLONE_PATH is missing")
	}

	// MySQL
	if os.Getenv("MYSQL_HOST") != "" {
		log.Printf("[HyperBackup] ✅ MySQL backup configured")
		configured++
	}

	// PostgreSQL
	if os.Getenv("POSTGRES_HOST") != "" {
		log.Printf("[HyperBackup] ✅ PostgreSQL backup configured")
		configured++
	}

	// MongoDB
	if os.Getenv("MONGO_HOST") != "" {
		log.Printf("[HyperBackup] ✅ MongoDB backup configured")
		configured++
	}

	// Traefik
	if os.Getenv("TRAEFIK_LOG_FILE") != "" {
		log.Printf("[HyperBackup] ✅ Traefik logrotate enabled")
		configured++
	}

	if useRseries == 0 {
		log.Printf("[HyperBackup] ⚠️ Warn: BACKUP WILL BE STORED LOCALLY ONLY")
	}

	if configured == 0 {
		log.Printf("[HyperBackup] 🤷 No backup services configured; nothing to do")
	}

	log.Printf("[HyperBackup] ✅ Configuration check complete")
	logDivider()
	return nil
}
func logDivider() {
	log.Println("════════════════════════════════════════════════════════════════")
}
