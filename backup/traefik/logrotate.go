// 📄backup/traefik/logrotate.go

package traefik

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func RotateAndBackup(logFile string) (string, int64, error) {
	info, err := os.Stat(logFile)
	if err != nil {
		return "", 0, fmt.Errorf("log file missing: %w", err)
	}
	if info.Size() == 0 {
		return "", 0, nil
	}

	timestamp := time.Now().Format("20060102_150405")
	baseDir := "/home/hyper-backup/traefik"
	filename := filepath.Base(logFile)
	rotated := fmt.Sprintf("%s.%s", filename, timestamp)
	rotatedPath := filepath.Join(baseDir, rotated)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", 0, fmt.Errorf("create base dir: %w", err)
	}
	n, err := copyFile(logFile, rotatedPath)
	if err != nil {
		return "", 0, fmt.Errorf("copy failed: %w", err)
	}

	if extra := os.Getenv("TRAEFIK_BACKUP_DIR"); extra != "" {
		extraPath := filepath.Join(extra, rotated)
		_ = os.MkdirAll(extra, 0755)
		_, _ = copyFile(logFile, extraPath)
	}

	_ = os.Remove(logFile)
	_, err = os.Create(logFile)
	if err != nil {
		return "", 0, fmt.Errorf("recreate log: %w", err)
	}

	return rotatedPath, n, nil
}

func copyFile(src, dst string) (int64, error) {
	in, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("open source file %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("create destination file %s: %w", dst, err)
	}
	defer out.Close()

	return io.Copy(out, in)
}
