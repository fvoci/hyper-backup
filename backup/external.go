package backup

import (
	"github.com/fvoci/hyper-backup/backup/folders"
	"github.com/fvoci/hyper-backup/backup/storage"
	"github.com/fvoci/hyper-backup/utilities"
)

// RunExternalBackups 함수는 폴더 백업을 수행한 후 rclone과 rsync를 이용해 외부 저장소로 데이터를 업로드합니다.
// 외부 백업 서비스 실행 중 오류가 발생하면 해당 오류를 반환합니다.
func RunExternalBackups() error {
	utilities.LogDivider()
	utilities.Logger.Info("☁️ [External Backups]")

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
