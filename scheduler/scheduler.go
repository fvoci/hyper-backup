package scheduler

import (
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/fvoci/hyper-backup/backup"
	utiles "github.com/fvoci/hyper-backup/utilities"
)

func Start(schedule string, interval string) {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = time.Now().Location().String()
	}
	utiles.Logger.Infof("[HyperBackup] 🌐 Timezone: %s", tz)

	switch {
	case schedule != "":
		startWithCron(schedule)
	case interval != "":
		startWithInterval(interval)
	default:
		utiles.Logger.Info("[HyperBackup] ⚠️ No schedule configured, defaulting to daily at midnight")
		startWithCron("0 0 * * *")
	}
}

func runBackupCycle(next time.Time) {
	start := time.Now()

	utiles.Logger.Info("🚀 [HyperBackup] Backup cycle started")
	utiles.Logger.Infof("🕒 %s", start.Format("2006-01-02 15:04:05"))

	backup.RunCoreServices()
	backup.RunExternalBackups()

	end := time.Now()
	utiles.Logger.Info("✅ [HyperBackup] Backup cycle completed")
	utiles.Logger.Infof("🕒 %s (Duration: %s)", end.Format("2006-01-02 15:04:05"), end.Sub(start).Round(time.Second))

	if !next.IsZero() {
		loc := next.Location()
		utiles.Logger.Infof("📅 Next backup scheduled at: %s (%s)", next.Format("2006-01-02 15:04:05"), loc.String())
	}

	utiles.LogDivider()
}

func startWithCron(schedule string) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		utiles.Logger.Fatalf("[HyperBackup] ❌ Invalid BACKUP_SCHEDULE '%s': %v", schedule, err)
	}

	next := sched.Next(time.Now().In(time.Local))
	utiles.Logger.Infof("[HyperBackup] 🔁 Scheduling backups with cron: \"%s\"", schedule)
	utiles.Logger.Infof("[HyperBackup] ⏳ Next backup at: %s (%s)", next.Format("2006-01-02 15:04:05"), next.Location().String())
	utiles.LogDivider()

	// cron 등록
	c := cron.New(cron.WithParser(parser), cron.WithLocation(time.Local))
	_, _ = c.AddFunc(schedule, func() {
		nextRun := sched.Next(time.Now().In(time.Local))
		runBackupCycle(nextRun)
	})

	// 초기 실행
	runBackupCycle(next)
	c.Start()

	select {}
}

func startWithInterval(hoursStr string) {
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

	// Create channel for termination signals
	terminationCh := make(chan os.Signal, 1)
	signal.Notify(terminationCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			next = time.Now().Add(dur)
			runBackupCycle(next)
		case <-terminationCh:
			utiles.Logger.Info("[HyperBackup] Received termination signal, shutting down...")
			return
		}
	}
}
