package main

import (
	"os"

	"github.com/fvoci/hyper-backup/scheduler"
	utiles "github.com/fvoci/hyper-backup/utilities"
)

func main() {
	utiles.Logger.Info("[HyperBackup] ⏱️ Backup process starting")
	utiles.LogDivider()

	if err := utiles.CheckConfig(); err != nil {
		utiles.Logger.Fatalf("[HyperBackup] ❌ Configuration check failed: %v", err)
	}

	schedule := os.Getenv("BACKUP_SCHEDULE")
	interval := os.Getenv("BACKUP_INTERVAL_HOURS")

	scheduler.Start(schedule, interval)
}
