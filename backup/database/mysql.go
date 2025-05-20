// 📄backup/mysql.go

package backup

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

func RunMySQL() {
	cfg, err := loadMySQLConfig()
	if err != nil {
		log.Printf("[MySQL] ❌ Configuration error: %v\n", err)
		return
	}

	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		log.Printf("[MySQL] ❌ Failed to create backup directory: %v\n", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql.gz", cfg.Database, timestamp)
	outputFile := filepath.Join(cfg.BackupDir, filename)

	log.Printf("[MySQL] 🐬 Backing up %s to %s\n", cfg.Database, outputFile)

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
		log.Printf("[MySQL] ❌ Failed to get dump stdout: %v\n", err)
		return
	}
	gzipCmd.Stdin = dumpOut

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Printf("[MySQL] ❌ Failed to create output file: %v\n", err)
		return
	}
	defer outFile.Close()
	gzipCmd.Stdout = outFile

	if err := dumpCmd.Start(); err != nil {
		log.Printf("[MySQL] ❌ mysqldump start error: %v\n", err)
		return
	}
	if err := gzipCmd.Start(); err != nil {
		log.Printf("[MySQL] ❌ gzip start error: %v\n", err)
		return
	}

	if err := dumpCmd.Wait(); err != nil {
		log.Printf("[MySQL] ❌ mysqldump execution error: %v\n", err)
	}
	if err := gzipCmd.Wait(); err != nil {
		log.Printf("[MySQL] ❌ gzip execution error: %v\n", err)
	}

	log.Printf("[MySQL] ✅ Backup completed successfully")
	log.Printf("\n")
}
