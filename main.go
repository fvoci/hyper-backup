package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fvoci/hyper-backup/backup"
	utiles "github.com/fvoci/hyper-backup/utilities"
	"github.com/robfig/cron/v3"
)

func main() {
	log.Println("[HyperBackup] ⏱️ Backup process starting")
	logDivider()
	if err := utiles.CheckConfig(); err != nil {
		log.Fatalf("[HyperBackup] ❌ Configuration check failed: %v", err)
	}

	schedule := os.Getenv("BACKUP_SCHEDULE")
	if schedule != "" {
		startWithCron(schedule)
		return
	}

	interval := os.Getenv("BACKUP_INTERVAL_HOURS")
	if interval != "" {
		startWithInterval(interval)
		return
	}

	log.Println("[HyperBackup] ⚠️ No schedule configured, defaulting to daily at midnight")
	startWithCron("0 0 * * *")
}

func runBackupCycle() {
	start := time.Now()

	log.Printf("🚀 [HyperBackup] Backup cycle started")
	log.Printf("🕒 %s", start.Format("2006-01-02 15:04:05"))

	// 1. Core service backups
	backup.RunCoreServices()

	// 2. External backups (folders + Rclone/Rsync)
	backup.RunExternalBackups()

	end := time.Now()
	logDivider()
	log.Printf("✅ [HyperBackup] Backup cycle completed")
	log.Printf("🕒 %s (Duration: %s)", end.Format("2006-01-02 15:04:05"), end.Sub(start).Round(time.Second))
	logDivider()
}

func logDivider() {
	log.Println("════════════════════════════════════════════════════════════════")
}

func startWithCron(schedule string) {
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(schedule)
	if err != nil {
		log.Fatalf("[HyperBackup] ❌ Invalid BACKUP_SCHEDULE '%s': %v", schedule, err)
	}
	next := sched.Next(time.Now().In(time.Local))
	log.Printf("[HyperBackup] 🔁 Scheduling backups with cron: \"%s\"", schedule)
	log.Printf("[HyperBackup] ⏳ Next backup at: %s", next.Format("2006-01-02 15:04:05"))
	logDivider()
	c := cron.New(cron.WithParser(parser), cron.WithLocation(time.Local))
	_, _ = c.AddFunc(schedule, runBackupCycle)

	runBackupCycle()
	c.Start()
	select {}
}

func startWithInterval(hoursStr string) {
	const fallback = 1
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours < 1 {
		log.Printf("[HyperBackup] ⚠️ Invalid BACKUP_INTERVAL_HOURS '%s'; using default %d시간", hoursStr, fallback)
		hours = fallback
	}

	dur := time.Duration(hours) * time.Hour
	log.Printf("[HyperBackup] 🔁 Scheduling backups every %d시간", hours)

	runBackupCycle()

	ticker := time.NewTicker(dur)
	defer ticker.Stop()

	for range ticker.C {
		runBackupCycle()
	}
}
