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
	utilities.LogDivider()

	if err := utilities.CheckConfig(); err != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Configuration check failed: %v", err)
	}

	// 컨텍스트 생성 및 종료 시그널 바인딩
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	schedule := os.Getenv("BACKUP_SCHEDULE")
	interval := os.Getenv("BACKUP_INTERVAL_HOURS")

	scheduler.StartWithContext(ctx, schedule, interval)
}
