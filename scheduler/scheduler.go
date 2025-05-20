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
// If neither is set, defaults to "0 0 * * *" (midnight).
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
