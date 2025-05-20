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
	if err := run(); err != nil {
		utilities.Logger.Error(err)
		os.Exit(1)
	}
}

func run() error {
	utilities.Logger.Info("[HyperBackup] ⏱️ Backup process starting")

	if err := utilities.CheckConfig(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schedule := os.Getenv("BACKUP_SCHEDULE")
	interval := os.Getenv("BACKUP_INTERVAL_HOURS")

	if schedule != "" && interval != "" {
		utilities.Logger.Warn("[HyperBackup] ⚠️ Both BACKUP_SCHEDULE and BACKUP_INTERVAL_HOURS are set. Using BACKUP_SCHEDULE.")
	}

	scheduler.StartWithContext(ctx, schedule, interval)
	return nil
}
