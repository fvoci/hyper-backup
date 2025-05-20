package storage

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fvoci/hyper-backup/utilities"
)

const (
	defaultRetentionDays = 14
	defaultRegion        = "us-east-1"
	backupDir            = "/home/hyper-backup"
)

type rcloneConfig struct {
	Remote    string
	Target    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	Retention int
}

func RunRclone() error {
	cfg, err := loadRcloneConfig()
	if err != nil {
		utilities.Logger.Errorf("[Rclone] ❌ Configuration error: %v", err)
		return err
	}

	if !waitForHTTP(cfg.Endpoint, 30*time.Second) {
		utilities.Logger.Error("[Rclone] ❌ S3 endpoint unreachable; skipping upload")
		return fmt.Errorf("endpoint unreachable: %s", cfg.Endpoint)
	}

	if err := cleanRemote(cfg); err != nil {
		utilities.Logger.Warnf("[Rclone] ⚠️ Remote cleanup error: %v", err)
	}

	if err := copyBackup(cfg); err != nil {
		utilities.Logger.Errorf("[Rclone] ❌ Upload failed: %v", err)
		return err
	}

	utilities.Logger.Info("[Rclone] ✅ Backup completed successfully")
	utilities.LogDivider()
	return nil
}

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

func waitForHTTP(url string, timeout time.Duration) bool {
	utilities.Logger.Infof("[Rclone] ⏳ Waiting for S3 endpoint %s", url)
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Head(url)
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			utilities.Logger.Info("[Rclone] ✅ Endpoint is reachable")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func cleanRemote(cfg *rcloneConfig) error {
	utilities.Logger.Infof("[Rclone] 🧹 Cleaning remote files older than %d days at %s", cfg.Retention, cfg.Target)
	age := fmt.Sprintf("%dd", cfg.Retention)
	cmdArgs := []string{"delete", cfg.Target, "--min-age", age}
	if cfgFile := os.Getenv("RCLONE_CONFIG_FILE"); cfgFile != "" {
		cmdArgs = append(cmdArgs, "--config", cfgFile)
	}
	cmd := exec.Command("rclone", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		utilities.Logger.Warnf("[Rclone] ⚠️ Remote cleanup failed: %v\nOutput:\n%s", err, out)
	}
	return err
}

func copyBackup(cfg *rcloneConfig) error {
	utilities.Logger.Infof("[Rclone] 🔄 Uploading %s to %s", backupDir, cfg.Target)
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
		utilities.Logger.Errorf("[Rclone] ❌ Upload error:\n%s", out)
	}
	return err
}
