package backup

import (
	"github.com/fvoci/hyper-backup/backup/folders"
	storage "github.com/fvoci/hyper-backup/backup/storage"
	utiles "github.com/fvoci/hyper-backup/utilities"
)

// RunExternalBackups runs folder compression and remote uploads via rclone/rsync.
func RunExternalBackups() error {
	utiles.LogDivider()
	utiles.Logger.Info("☁️ [External Backups]")

	// Folder backup doesn't return error (yet), so just run it first
	folders.RunFileBackup()

	services := []service{
		{
			Name:     "Rclone",
			EnvKeys:  []string{"RCLONE_REMOTE", "RCLONE_PATH"},
			RunFunc:  storage.RunRclone,
			Optional: true,
		},
		{
			Name:     "Rsync",
			EnvKeys:  []string{"RSYNC_SRC", "RSYNC_DEST"},
			RunFunc:  storage.RunRsync,
			Optional: true,
		},
	}

	return runServices(services)
}
