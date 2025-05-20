package backup

import (
	"github.com/fvoci/hyper-backup/backup/folders"
	storage "github.com/fvoci/hyper-backup/backup/storage"
	utiles "github.com/fvoci/hyper-backup/utilities"
)

func RunExternalBackups() {
	utiles.LogDivider()
	utiles.Logger.Info("☁️ [External Backups]")

	// Step 1: folder compression
	folders.RunFileBackup()

	// Step 2: sync via Rclone / Rsync
	services := []service{
		{"Rclone", []string{"RCLONE_REMOTE", "RCLONE_PATH"}, storage.RunRclone, true},
		{"Rsync", []string{"RSYNC_SRC", "RSYNC_DEST"}, storage.RunRsync, true},
	}
	runServices(services)
}
