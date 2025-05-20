package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/fvoci/hyper-backup/scheduler"
	"github.com/fvoci/hyper-backup/utilities"
)

// main은 백업 프로세스를 시작하고, 오류 발생 시 로그를 남기고 비정상 종료합니다.
func main() {
	if err := run(); err != nil {
		utilities.Logger.Error(err)
		os.Exit(1)
	}
}

// run은 백업 프로세스를 초기화하고, 환경 변수에 따라 스케줄러를 설정하여 백업 작업을 시작합니다.
// 설정 검증에 실패하면 오류를 반환하며, OS 종료 신호를 감지하여 백업 프로세스의 정상 종료를 지원합니다.
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
