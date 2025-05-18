// backup/rclone.go

package backup

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	defaultRetentionDays = 14
	defaultRegion        = "us-west-1"
	backupDir            = "/home/hyper-backup"
)

// rcloneConfig holds configuration for rclone backup.
type rcloneConfig struct {
	Remote    string
	Target    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	Retention int
}

// RunRclone executes the rclone backup workflow:
// 1) Wait for S3 endpoint readiness
// 2) Clean old local files
// 3) Clean old remote objects
// 4) Upload backup directory via rclone
func RunRclone() {
	cfg, err := loadRcloneConfig()
	if err != nil {
		fmt.Printf("[Rclone] ❌ Configuration error: %v\n", err)
		return
	}

	if !waitForHTTP(cfg.Endpoint, 30*time.Second) {
		fmt.Println("[Rclone] ❌ S3 endpoint unreachable; skipping upload")
		return
	}

	// if err := cleanLocal(backupDir, cfg.Retention); err != nil {
	// 	fmt.Printf("[Rclone] ⚠️ Local cleanup error: %v\n", err)
	// }

	if err := cleanRemote(cfg); err != nil {
		fmt.Printf("[Rclone] ⚠️ Remote cleanup error: %v\n", err)
	}

	if err := copyBackup(cfg); err != nil {
		fmt.Printf("[Rclone] ❌ Upload failed: %v\n", err)
		return
	}

	fmt.Println("[Rclone] ✅ Backup completed successfully")
}

// loadRcloneConfig reads and validates environment variables for rclone.
func loadRcloneConfig() (*rcloneConfig, error) {
	remote := os.Getenv("RCLONE_REMOTE")
	target := os.Getenv("RCLONE_PATH")
	endpoint := os.Getenv("S3_ENDPOINT")
	if remote == "" || target == "" || endpoint == "" {
		return nil, fmt.Errorf("RCLONE_REMOTE, RCLONE_PATH and S3_ENDPOINT must be set")
	}

	retention := defaultRetentionDays
	if str := os.Getenv("RCLONE_RETENTION_DAYS"); str != "" {
		if v, err := strconv.Atoi(str); err == nil && v > 0 {
			retention = v
		}
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultRegion
	}

	return &rcloneConfig{
		Remote:    remote,
		Target:    target,
		Endpoint:  endpoint,
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:    region,
		Retention: retention,
	}, nil
}

// waitForHTTP polls the given URL until it responds with status <500 or timeout.
func waitForHTTP(url string, timeout time.Duration) bool {
	fmt.Printf("[Rclone] ⏳ Waiting for S3 endpoint %s\n", url)
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Head(url)
		if err == nil && resp.StatusCode < 500 {
			fmt.Println("[Rclone] ✅ Endpoint is reachable")
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

// cleanLocal removes files older than the retention period from the local backup directory.
// func cleanLocal(path string, days int) error {
// 	cutoff := time.Now().AddDate(0, 0, -days)
// 	fmt.Printf("[Rclone] 🧹 Cleaning local files older than %d days in %s\n", days, path)

// 	return filepath.Walk(path, func(fp string, info os.FileInfo, err error) error {
// 		if err != nil || info.IsDir() {
// 			return err
// 		}
// 		if info.ModTime().Before(cutoff) {
// 			fmt.Printf("[Rclone] 🗑️ Deleting local file: %s (modified: %s)\n",
// 				fp, info.ModTime().Format(time.RFC3339))
// 			return os.Remove(fp)
// 		}
// 		return nil
// 	})
// }

// cleanRemote deletes remote objects older than the retention period from the remote target.
func cleanRemote(cfg *rcloneConfig) error {
	fmt.Printf("[Rclone] 🧹 Cleaning remote files older than %d days at %s\n", cfg.Retention, cfg.Target)
	age := fmt.Sprintf("%dd", cfg.Retention)
	cmdArgs := []string{"delete", cfg.Target, "--min-age", age}
	if cfgFile := os.Getenv("RCLONE_CONFIG_FILE"); cfgFile != "" {
		cmdArgs = append(cmdArgs, "--config", cfgFile)
	}
	cmd := exec.Command("rclone", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Rclone] ⚠️ Remote cleanup failed: %v\nOutput:\n%s", err, out)
	}
	return err
}

// copyBackup uploads the local backup directory to the remote target via rclone.
func copyBackup(cfg *rcloneConfig) error {
	fmt.Printf("[Rclone] 🔄 Uploading %s to %s\n", backupDir, cfg.Target)
	key := strings.ToUpper(cfg.Remote)

	env := os.Environ()

	env = append(env,
		"RCLONE_CONFIG_"+key+"_TYPE=s3",
		"RCLONE_CONFIG_"+key+"_PROVIDER=Other",
		"RCLONE_CONFIG_"+key+"_ACCESS_KEY_ID="+cfg.AccessKey,
		"RCLONE_CONFIG_"+key+"_SECRET_ACCESS_KEY="+cfg.SecretKey,
		"RCLONE_CONFIG_"+key+"_ENDPOINT="+cfg.Endpoint,
		"RCLONE_CONFIG_"+key+"_REGION="+cfg.Region,
		"RCLONE_CONFIG_"+key+"_ENV_AUTH=false",
	)

	cmd := exec.Command("rclone", "copy", backupDir, cfg.Target)
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("[Rclone] ❌ Upload error:\n%s\n", out)
	}
	return err
}
