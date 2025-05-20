package scheduler

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/fvoci/hyper-backup/backup"
	utiles "github.com/fvoci/hyper-backup/utilities"
	"github.com/robfig/cron/v3"
)

func StartWithContext(ctx context.Context, schedule string, interval string) {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = time.Now().Location().String()
	}
	utiles.Logger.Infof("[HyperBackup] 🌐 Timezone: %s", tz)

	switch {
	case schedule != "":
		startWithCronContext(ctx, schedule)
	case interval != "":
		startWithIntervalContext(ctx, interval)
	default:
		utiles.Logger.Info("[HyperBackup] ⚠️ No schedule configured, defaulting to daily at midnight")
		startWithCronContext(ctx, "0 0 * * *")
	}
}

func runBackupCycle(next time.Time) {
	start := time.Now()

	utiles.Logger.Info("🚀 [HyperBackup] Backup cycle started")
	utiles.Logger.Infof("🕒 %s", start.Format("2006-01-02 15:04:05"))

	if err := backup.RunCoreServices(); err != nil {
		utiles.Logger.Errorf("[HyperBackup] ❌ Core service backup failed: %v", err)
	}

	if err := backup.RunExternalBackups(); err != nil {
		utiles.Logger.Errorf("[HyperBackup] ❌ External backup failed: %v", err)
	}

	end := time.Now()
	utiles.Logger.Info("✅ [HyperBackup] Backup cycle completed")
	utiles.Logger.Infof("🕒 %s (Duration: %s)", end.Format("2006-01-02 15:04:05"), end.Sub(start).Round(time.Second))

	if !next.IsZero() {
		utiles.Logger.Infof("📅 Next backup scheduled at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location().String())
	}

	utiles.LogDivider()
}

func startWithCronContext(ctx context.Context, schedule string) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		utiles.Logger.Fatalf("[HyperBackup] ❌ Invalid BACKUP_SCHEDULE '%s': %v", schedule, err)
	}

	next := sched.Next(time.Now().In(time.Local))
	utiles.Logger.Infof("[HyperBackup] 🔁 Scheduling backups with cron: \"%s\"", schedule)
	utiles.Logger.Infof("[HyperBackup] ⏳ Next backup at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location().String())
	utiles.LogDivider()

	c := cron.New(cron.WithParser(parser), cron.WithLocation(time.Local))
	_, _ = c.AddFunc(schedule, func() {
		nextRun := sched.Next(time.Now().In(time.Local))
		runBackupCycle(nextRun)
	})

	c.Start()
	runBackupCycle(next)

	<-ctx.Done()
	utiles.Logger.Info("[HyperBackup] 🛑 Received shutdown signal. Stopping cron scheduler...")
	c.Stop()
	utiles.Logger.Info("[HyperBackup] ✅ Scheduler stopped. Goodbye.")
}

func startWithIntervalContext(ctx context.Context, hoursStr string) {
	const fallback = 1
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		utiles.Logger.Warnf("[HyperBackup] ⚠️ Invalid BACKUP_INTERVAL_HOURS '%s'; using default %d hour(s)", hoursStr, fallback)
		hours = fallback
	}

	dur := time.Duration(hours) * time.Hour
	utiles.Logger.Infof("[HyperBackup] 🔁 Scheduling backups every %d hour(s)", hours)
	utiles.LogDivider()

	next := time.Now().Add(dur)
	runBackupCycle(next)

	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			next = time.Now().Add(dur)
			runBackupCycle(next)
		case <-ctx.Done():
			utiles.Logger.Info("[HyperBackup] 🛑 Received shutdown signal. Stopping interval scheduler...")
			utiles.Logger.Info("[HyperBackup] ✅ Scheduler stopped. Goodbye.")
			return
		}
	}
}
