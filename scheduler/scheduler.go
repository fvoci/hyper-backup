package scheduler

import (
	"context"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/fvoci/hyper-backup/backup"
	"github.com/fvoci/hyper-backup/utilities"
	"github.com/robfig/cron/v3"
)

// StartWithContext starts the backup scheduler.
// ctx: for graceful shutdown via cancellation.
// schedule: cron expression (e.g., "0 0 * * *") — takes priority.
// interval: hour-based string (e.g., "24") if cron is not used.
// StartWithContext는 주어진 컨텍스트와 스케줄 설정에 따라 백업 스케줄러를 시작합니다.
//
// 'schedule'이 크론 표현식으로 지정되면 크론 기반 스케줄러를, 'interval'이 지정되면 고정 간격 스케줄러를 사용합니다.
// 둘 다 지정되지 않은 경우 매일 자정(크론 "0 0 * * *")에 백업을 실행합니다.
// 'TZ' 환경 변수가 설정되어 있으면 해당 타임존을 적용하며, 유효하지 않을 경우 시스템 기본 타임존을 사용합니다.
// 컨텍스트가 취소되면 스케줄러가 정상적으로 종료됩니다.
func StartWithContext(ctx context.Context, schedule, interval string) {
	if tzName := os.Getenv("TZ"); tzName != "" {
		if loc, err := time.LoadLocation(tzName); err == nil {
			time.Local = loc
		} else {
			utilities.Logger.Warnf("[HyperBackup] ⚠️ Invalid TZ '%s', using system default: %v", tzName, err)
		}
	}
	utilities.Logger.Infof("[HyperBackup] 🌐 Timezone: %s", time.Local.String())

	switch {
	case schedule != "":
		startWithCron(ctx, schedule)
	case interval != "":
		startWithInterval(ctx, interval)
	default:
		utilities.Logger.Info("[HyperBackup] ⚠️ No schedule set. Defaulting to daily at midnight.")
		startWithCron(ctx, "0 0 * * *")
	}
}

// runBackupCycle는 백업 주기를 실행하고, 시작 및 종료 시간, 소요 시간, 다음 예약 백업 시간을 로그로 기록합니다.
// 내부적으로 핵심 서비스와 외부 백업을 순차적으로 실행하며, 오류 발생 시 로그에 남깁니다.
func runBackupCycle(next time.Time) {
	start := time.Now()

	utilities.Logger.Info("🚀 [HyperBackup] Backup cycle started")
	utilities.Logger.Infof("🕒 %s", start.Format("2006-01-02 15:04:05"))

	if err := backup.RunCoreServices(); err != nil {
		utilities.Logger.Errorf("[HyperBackup] ❌ Core services failed: %v", err)
	}
	if err := backup.RunExternalBackups(); err != nil {
		utilities.Logger.Errorf("[HyperBackup] ❌ External backups failed: %v", err)
	}

	end := time.Now()
	utilities.Logger.Info("✅ [HyperBackup] Backup cycle completed")
	utilities.Logger.Infof("🕒 %s (Duration: %s)", end.Format("2006-01-02 15:04:05"), end.Sub(start).Round(time.Second))

	if !next.IsZero() {
		utilities.Logger.Infof("📅 Next backup at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location())
	}
	utilities.LogDivider()
}

// startWithCron는 주어진 cron 스케줄 문자열에 따라 백업 작업을 예약하고, 컨텍스트가 취소될 때까지 주기적으로 백업을 실행합니다.
// 이전 백업이 아직 실행 중인 경우 중복 실행을 방지하며, 스케줄 파싱 오류 시 프로그램을 종료합니다.
func startWithCron(ctx context.Context, schedule string) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	spec, err := parser.Parse(schedule)
	if err != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Invalid BACKUP_SCHEDULE '%s': %v", schedule, err)
	}

	var running int32
	next := spec.Next(time.Now().In(time.Local))
	utilities.Logger.Infof("[HyperBackup] 🔁 Using cron: \"%s\"", schedule)
	utilities.Logger.Infof("[HyperBackup] ⏳ Next backup at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location())
	utilities.LogDivider()

	c := cron.New(
		cron.WithParser(parser),
		cron.WithLocation(time.Local),
		cron.WithChain(
			cron.Recover(cron.PrintfLogger(utilities.Logger)),
		),
	)

	if _, err := c.AddFunc(schedule, func() {
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			utilities.Logger.Warn("[HyperBackup] ⚠️ Previous backup still running. Skipping this cycle.")
			return
		}
		defer atomic.StoreInt32(&running, 0)
		runBackupCycle(spec.Next(time.Now().In(time.Local)))
	}); err != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Failed to schedule job: %v", err)
	}

	c.Start()
	runBackupCycle(next)

	<-ctx.Done()
	utilities.Logger.Info("[HyperBackup] 🛑 Stopping cron scheduler...")
	c.Stop()
	utilities.Logger.Info("[HyperBackup] ✅ Scheduler stopped")
}

// startWithInterval는 지정된 시간 간격(시간 단위)마다 백업 작업을 실행하는 스케줄러를 시작합니다.
// 컨텍스트가 취소되면 스케줄러를 안전하게 종료합니다.
// 잘못된 간격 입력 시 기본값(1시간)을 사용하며, 이전 백업이 완료되지 않은 경우 중복 실행을 방지합니다.
func startWithInterval(ctx context.Context, hoursStr string) {
	const fallback = 1
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		utilities.Logger.Warnf("[HyperBackup] ⚠️ Invalid BACKUP_INTERVAL_HOURS '%s'. Using default %d hour(s)", hoursStr, fallback)
		hours = fallback
	}

	dur := time.Duration(hours) * time.Hour
	utilities.Logger.Infof("[HyperBackup] 🔁 Using interval: every %d hour(s)", hours)
	utilities.LogDivider()

	var running int32
	next := time.Now().Add(dur)
	runBackupCycle(next)

	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !atomic.CompareAndSwapInt32(&running, 0, 1) {
				utilities.Logger.Warn("[HyperBackup] ⚠️ Previous backup still running. Skipping this cycle.")
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						utilities.Logger.Errorf("[HyperBackup] ❌ Panic during backup: %v", r)
					}
					atomic.StoreInt32(&running, 0)
				}()
				runBackupCycle(time.Now().Add(dur))
			}()

		case <-ctx.Done():
			utilities.Logger.Info("[HyperBackup] 🛑 Stopping interval scheduler...")
			utilities.Logger.Info("[HyperBackup] ✅ Scheduler stopped")
			return
		}
	}
}
