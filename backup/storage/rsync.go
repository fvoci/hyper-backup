package storage

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fvoci/hyper-backup/utilities"
)

type rsyncConfig struct {
	Src  string
	Dest string
}

// loadRsyncConfig는 환경 변수 RSYNC_SRC와 RSYNC_DEST에서 소스 및 대상 경로를 읽어 rsyncConfig를 반환합니다.
// 두 환경 변수 중 하나라도 설정되어 있지 않으면 오류를 반환합니다.
func loadRsyncConfig() (*rsyncConfig, error) {
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	if src == "" || dest == "" {
		return nil, fmt.Errorf("RSYNC_SRC and RSYNC_DEST must be set")
	}
	return &rsyncConfig{Src: src, Dest: dest}, nil
}
// RunRsync는 환경 변수로 지정된 소스와 목적지 디렉터리를 사용하여 로컬 디렉터리 백업을 수행한다.
// rsync 명령어를 실행하여 소스 디렉터리의 내용을 목적지로 동기화하며, 동기화 과정에서 오류가 발생하면 에러를 반환한다.
func RunRsync() error {
	cfg, err := loadRsyncConfig()
	if err != nil {
		utilities.Logger.Errorf("[Rsync] ❌ Configuration error: %v", err)
		return err
	}

	utilities.Logger.Infof("[Rsync] 📁 Backing up local directory: %s → %s", cfg.Src, cfg.Dest)

	if err := os.MkdirAll(cfg.Dest, 0755); err != nil {
		utilities.Logger.Errorf("[Rsync] ❌ Failed to create destination directory: %v", err)
		return err
	}

	cmd := exec.Command("rsync", "-a", "--delete", cfg.Src+"/", cfg.Dest+"/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		utilities.Logger.Errorf("[Rsync] ❌ rsync execution failed: %v\nOutput:\n%s", err, string(output))
		return err
	}

	utilities.Logger.Info("[Rsync] ✅ Local backup completed successfully")
	utilities.LogDivider()
	return nil
}
