// main.go

package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fvoci/hyper-backup/backup"
	"github.com/fvoci/hyper-backup/utiles"
)

// main is the entry point for the HyperBackup application.
// It checks configuration, runs an initial backup, and then schedules recurring backups.
func main() {
	fmt.Println("[HyperBackup] ⏱️ Backup process starting")

	// Validate essential environment variables
	if err := utiles.CheckConfig(); err != nil {
		fmt.Printf("[HyperBackup] ❌ Configuration check failed: %v\n", err)
		os.Exit(1)
	}

	interval := getBackupInterval()
	fmt.Printf("[HyperBackup] 🔁 Scheduling backups every %d seconds\n", interval)

	// Perform an immediate backup on startup
	fmt.Println("[HyperBackup] 🚀 Running initial backup")
	backup.RunAll()

	// Set up ticker for periodic backups
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("[HyperBackup] 🔁 Interval reached; running scheduled backup")
		backup.RunAll()
	}
}

// getBackupInterval parses BACKUP_INTERVAL environment variable (in seconds).
// Defaults to 3600 seconds (1 hour) if unset or invalid.
func getBackupInterval() int {
	const defaultInterval = 3600
	intervalStr := os.Getenv("BACKUP_INTERVAL")
	if intervalStr == "" {
		return defaultInterval
	}

	sec, err := strconv.Atoi(intervalStr)
	if err != nil || sec < 1 {
		fmt.Printf("[HyperBackup] ⚠️ Invalid BACKUP_INTERVAL '%s'; using default %d seconds\n", intervalStr, defaultInterval)
		return defaultInterval
	}
	return sec
}
