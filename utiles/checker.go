// utilises/checker.go

package utiles

import (
	"fmt"
	"os"
)

// CheckConfig inspects essential environment variables for each backup service.
// It prints status messages and returns an error if no backup service is configured.
func CheckConfig() error {
	fmt.Println("[Checker] 🔍 Checking environment variables")

	var enabled int

	// rsync configuration
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	switch {
	case src != "" && dest != "":
		fmt.Println("[Checker] ✅ rsync backup configured")
		enabled++
	case src != "" || dest != "":
		fmt.Println("[Checker] ⚠️ rsync: RSYNC_SRC or RSYNC_DEST is missing")
	}

	// rclone configuration
	remote := os.Getenv("RCLONE_REMOTE")
	path := os.Getenv("RCLONE_PATH")
	switch {
	case remote != "" && path != "":
		fmt.Println("[Checker] ✅ rclone backup configured")
		enabled++
	case remote != "" || path != "":
		fmt.Println("[Checker] ⚠️ rclone: RCLONE_REMOTE or RCLONE_PATH is missing")
	}

	// MySQL configuration
	if os.Getenv("MYSQL_HOST") != "" {
		fmt.Println("[Checker] ✅ MySQL backup configured")
		enabled++
	}

	// PostgreSQL configuration
	if os.Getenv("POSTGRES_HOST") != "" {
		fmt.Println("[Checker] ✅ PostgreSQL backup configured")
		enabled++
	}

	// MongoDB configuration
	if os.Getenv("MONGO_HOST") != "" {
		fmt.Println("[Checker] ✅ MongoDB backup configured")
		enabled++
	}

	if enabled == 0 {
		return fmt.Errorf("no backup service configured; please set at least one of RSYNC, RCLONE, MYSQL, POSTGRES or MONGO variables")
	}

	fmt.Println("[Checker] ✅ Configuration check complete")
	return nil
}
