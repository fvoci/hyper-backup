package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fvoci/hyper-backup/scheduler"
	"github.com/fvoci/hyper-backup/utilities"
)

func main() {
	utilities.Logger.Info("[HyperBackup] ⏱️ Backup process starting")

	if err := utilities.CheckConfig(); err != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Configuration check failed: %v", err)
	}

	// Create context and bind shutdown signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schedule := os.Getenv("BACKUP_SCHEDULE")
	interval := os.Getenv("BACKUP_INTERVAL_HOURS")

	if schedule != "" && interval != "" {
		utilities.Logger.Warn("[HyperBackup] ⚠️ Both BACKUP_SCHEDULE and BACKUP_INTERVAL_HOURS set; BACKUP_SCHEDULE takes precedence.")
	}

	scheduler.StartWithContext(ctx, schedule, interval)
}
