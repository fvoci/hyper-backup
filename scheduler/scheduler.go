package scheduler

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/fvoci/hyper-backup/backup"
	"github.com/fvoci/hyper-backup/utilities"
	"github.com/robfig/cron/v3"
)

func StartWithContext(ctx context.Context, schedule string, interval string) {
	tzEnv := os.Getenv("TZ")
	if tzEnv != "" {
		if loc, err := time.LoadLocation(tzEnv); err == nil {
			time.Local = loc
		} else {
			utilities.Logger.Warnf("[HyperBackup] ⚠️ Invalid TZ '%s', falling back to system timezone: %v", tzEnv, err)
		}
	}
	tz := time.Local.String()
	utilities.Logger.Infof("[HyperBackup] 🌐 Timezone: %s", tz)

	switch {
	case schedule != "":
		startWithCronContext(ctx, schedule)
	case interval != "":
		startWithIntervalContext(ctx, interval)
	default:
		utilities.Logger.Info("[HyperBackup] ⚠️ No schedule configured, defaulting to daily at midnight")
		startWithCronContext(ctx, "0 0 * * *")
	}
}

func runBackupCycle(next time.Time) {
	start := time.Now()

	utilities.Logger.Info("🚀 [HyperBackup] Backup cycle started")
	utilities.Logger.Infof("🕒 %s", start.Format("2006-01-02 15:04:05"))

	if err := backup.RunCoreServices(); err != nil {
		utilities.Logger.Errorf("[HyperBackup] ❌ Core service backup failed: %v", err)
	}

	if err := backup.RunExternalBackups(); err != nil {
		utilities.Logger.Errorf("[HyperBackup] ❌ External backup failed: %v", err)
	}

	end := time.Now()
	utilities.Logger.Info("✅ [HyperBackup] Backup cycle completed")
	utilities.Logger.Infof("🕒 %s (Duration: %s)", end.Format("2006-01-02 15:04:05"), end.Sub(start).Round(time.Second))

	if !next.IsZero() {
		utilities.Logger.Infof("📅 Next backup scheduled at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location().String())
	}

	utilities.LogDivider()
}

func startWithCronContext(ctx context.Context, schedule string) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Invalid BACKUP_SCHEDULE '%s': %v", schedule, err)
	}

	next := sched.Next(time.Now().In(time.Local))
	utilities.Logger.Infof("[HyperBackup] 🔁 Scheduling backups with cron: \"%s\"", schedule)
	utilities.Logger.Infof("[HyperBackup] ⏳ Next backup at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location().String())
	utilities.LogDivider()

	c := cron.New(cron.WithParser(parser), cron.WithLocation(time.Local))
	if _, addErr := c.AddFunc(schedule, func() {
		nextRun := sched.Next(time.Now().In(time.Local))
		runBackupCycle(nextRun)
	}); addErr != nil {
		utilities.Logger.Fatalf("[HyperBackup] ❌ Failed to add cron job: %v", addErr)
	}

	c.Start()
	runBackupCycle(next)

	<-ctx.Done()
	utilities.Logger.Info("[HyperBackup] 🛑 Received shutdown signal. Stopping cron scheduler...")
	c.Stop()
	utilities.Logger.Info("[HyperBackup] ✅ Scheduler stopped. Goodbye.")
}

func startWithIntervalContext(ctx context.Context, hoursStr string) {
	const fallback = 1
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		utilities.Logger.Warnf("[HyperBackup] ⚠️ Invalid BACKUP_INTERVAL_HOURS '%s'; using default %d hour(s)", hoursStr, fallback)
		hours = fallback
	}

	dur := time.Duration(hours) * time.Hour
	utilities.Logger.Infof("[HyperBackup] 🔁 Scheduling backups every %d hour(s)", hours)
	utilities.LogDivider()

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
			utilities.Logger.Info("[HyperBackup] 🛑 Received shutdown signal. Stopping interval scheduler...")
			utilities.Logger.Info("[HyperBackup] ✅ Scheduler stopped. Goodbye.")
			return
		}
	}
}
