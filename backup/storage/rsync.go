// 📄backup/rsync.go

package backup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
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

func RunRsync() {
	cfg, err := loadRsyncConfig()
	if err != nil {
		log.Printf("[Rsync] ❌ Configuration error: %v\n", err)
		return
	}

	log.Printf("[Rsync] 📁 Backing up local directory: %s → %s\n", cfg.Src, cfg.Dest)

	if err := os.MkdirAll(cfg.Dest, 0755); err != nil {
		log.Printf("[Rsync] ❌ Failed to create destination directory: %v\n", err)
		return
	}

	cmd := exec.Command("rsync", "-a", "--delete", cfg.Src+"/", cfg.Dest+"/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[Rsync] ❌ rsync execution failed: %v\nOutput:\n%s\n", err, string(output))
		return
	}

	log.Printf("[Rsync] ✅ Local backup completed successfully")
	log.Printf("\n")
}
