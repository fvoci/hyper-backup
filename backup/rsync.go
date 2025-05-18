// backup/rsync.go

package backup

import (
	"fmt"
	"os"
	"os/exec"
)

// rsyncConfig holds configuration for rsync backup.
type rsyncConfig struct {
	Src  string
	Dest string
}

// loadRsyncConfig reads and validates RSYNC_SRC and RSYNC_DEST environment variables.
func loadRsyncConfig() (*rsyncConfig, error) {
	src := os.Getenv("RSYNC_SRC")
	dest := os.Getenv("RSYNC_DEST")
	if src == "" || dest == "" {
		return nil, fmt.Errorf("RSYNC_SRC and RSYNC_DEST must be set")
	}
	return &rsyncConfig{Src: src, Dest: dest}, nil
}

// RunRsync performs a local directory backup using rsync.
func RunRsync() {
	cfg, err := loadRsyncConfig()
	if err != nil {
		fmt.Printf("[Rsync] ❌ Configuration error: %v\n", err)
		return
	}

	fmt.Printf("[Rsync] 📁 Backing up local directory: %s → %s\n", cfg.Src, cfg.Dest)

	// Ensure destination directory exists
	if err := os.MkdirAll(cfg.Dest, 0755); err != nil {
		fmt.Printf("[Rsync] ❌ Failed to create destination directory: %v\n", err)
		return
	}

	// Execute rsync with archive mode and deletion of extraneous files
	cmd := exec.Command("rsync", "-a", "--delete", cfg.Src+"/", cfg.Dest+"/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Rsync] ❌ rsync execution failed: %v\nOutput:\n%s\n", err, string(output))
		return
	}

	fmt.Println("[Rsync] ✅ Local backup completed successfully")
}
