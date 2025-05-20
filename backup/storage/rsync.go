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

func loadRsyncConfig() (*rsyncConfig, error) {
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	if src == "" || dest == "" {
		return nil, fmt.Errorf("RSYNC_SRC and RSYNC_DEST must be set")
	}
	return &rsyncConfig{Src: src, Dest: dest}, nil
}
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
