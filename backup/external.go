// 📄 backup/external.go

package backup

import (
	"log"

	"github.com/fvoci/hyper-backup/backup/folders"
	storage "github.com/fvoci/hyper-backup/backup/storage"
)

func RunExternalBackups() {
	log.Printf("\n")
	log.Printf("☁️ [External Backups]")

	// Step 1: folder compression
	folders.RunFileBackup()

	// Step 2: sync via Rclone / Rsync
	services := []service{
		{"Rclone", []string{"RCLONE_REMOTE", "RCLONE_PATH"}, storage.RunRclone, true},
		{"Rsync", []string{"RSYNC_SRC", "RSYNC_DEST"}, storage.RunRsync, true},
	}
	runServices(services)
}
