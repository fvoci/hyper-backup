package backup

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fvoci/hyper-backup/utilities"
)

type mysqlConfig struct {
	Host      string
	Port      string
	User      string
	Password  string
	Database  string
	BackupDir string
}

func loadMySQLConfig() (*mysqlConfig, error) {
	dsn := os.Getenv("MYSQL_DSN")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASSWORD")
	db := os.Getenv("MYSQL_DATABASE")
	backupDir := os.Getenv("MYSQL_BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/home/hyper-backup/mysql"
	}

	if dsn != "" {
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid MYSQL_DSN: %v", err)
		}
		if u.User != nil {
			user = u.User.Username()
			if pwd, ok := u.User.Password(); ok {
				pass = pwd
			}
		}
		hostPort := u.Host
		if h, p, err := net.SplitHostPort(hostPort); err == nil {
			host = h
			port = p
		} else {
			host = hostPort
		}
		if name := strings.TrimPrefix(u.Path, "/"); name != "" {
			db = name
		}
	}

	if host == "" || user == "" || pass == "" || db == "" {
		return nil, fmt.Errorf("MYSQL_HOST, MYSQL_USER, MYSQL_PASSWORD and MYSQL_DATABASE must be set")
	}

	return &mysqlConfig{
		Host:      host,
		Port:      port,
		User:      user,
		Password:  pass,
		Database:  db,
		BackupDir: backupDir,
	}, nil
}

func RunMySQL() error {
	cfg, err := loadMySQLConfig()
	if err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ Configuration error: %v", err)
		return err
	}

	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ Failed to create backup directory: %v", err)
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql.gz", cfg.Database, timestamp)
	outputFile := filepath.Join(cfg.BackupDir, filename)

	utilities.Logger.Infof("[MySQL] 🐬 Backing up %s to %s", cfg.Database, outputFile)

	dumpArgs := []string{
		"-h", cfg.Host,
		"-P", cfg.Port,
		"-u", cfg.User,
		fmt.Sprintf("-p%s", cfg.Password),
		cfg.Database,
	}
	dumpCmd := exec.Command("mysqldump", dumpArgs...)

	gzipCmd := exec.Command("gzip")
	dumpOut, err := dumpCmd.StdoutPipe()
	if err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ Failed to get dump stdout: %v", err)
		return err
	}
	gzipCmd.Stdin = dumpOut

	outFile, err := os.Create(outputFile)
	if err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ Failed to create output file: %v", err)
		return err
	}
	defer outFile.Close()
	gzipCmd.Stdout = outFile

	if err := dumpCmd.Start(); err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ mysqldump start error: %v", err)
		return err
	}
	if err := gzipCmd.Start(); err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ gzip start error: %v", err)
		return err
	}

	if err := dumpCmd.Wait(); err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ mysqldump execution error: %v", err)
		return err
	}
	if err := gzipCmd.Wait(); err != nil {
		utilities.Logger.Errorf("[MySQL] ❌ gzip execution error: %v", err)
		return err
	}

	utilities.Logger.Info("[MySQL] ✅ Backup completed successfully")
	utilities.LogDivider()
	return nil
}
