// backup/mysql.go

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
)

// mysqlConfig holds MySQL backup configuration loaded from environment variables.
type mysqlConfig struct {
	Host      string
	Port      string
	User      string
	Password  string
	Database  string
	BackupDir string
}

// loadMySQLConfig reads and validates MySQL backup settings.
// Supports full DSN via MYSQL_DSN or individual vars.
func loadMySQLConfig() (*mysqlConfig, error) {
	// Read environment variables
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

	// Parse DSN if provided
	if dsn != "" {
		u, err := url.Parse(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid MYSQL_DSN: %v", err)
		}
		// Extract user and password
		if u.User != nil {
			user = u.User.Username()
			if pwd, ok := u.User.Password(); ok {
				pass = pwd
			}
		}
		// Extract host and port
		hostPort := u.Host
		if h, p, err := net.SplitHostPort(hostPort); err == nil {
			host = h
			port = p
		} else {
			host = hostPort
		}
		// Extract database name from path
		if name := strings.TrimPrefix(u.Path, "/"); name != "" {
			db = name
		}
	}

	// Validate required fields
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

// RunMySQL performs a compressed dump of the specified MySQL database.
func RunMySQL() {
	cfg, err := loadMySQLConfig()
	if err != nil {
		fmt.Printf("[MySQL] ❌ Configuration error: %v\n", err)
		return
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		fmt.Printf("[MySQL] ❌ Failed to create backup directory: %v\n", err)
		return
	}

	// Prepare output file name with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.sql.gz", cfg.Database, timestamp)
	outputFile := filepath.Join(cfg.BackupDir, filename)

	fmt.Printf("[MySQL] 🐬 Backing up %s to %s\n", cfg.Database, outputFile)

	// Build mysqldump command arguments
	dumpArgs := []string{
		"-h", cfg.Host,
		"-P", cfg.Port,
		"-u", cfg.User,
		fmt.Sprintf("-p%s", cfg.Password),
		cfg.Database,
	}
	dumpCmd := exec.Command("mysqldump", dumpArgs...)

	// Pipe to gzip
	gzipCmd := exec.Command("gzip")
	dumpOut, err := dumpCmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[MySQL] ❌ Failed to get dump stdout: %v\n", err)
		return
	}
	gzipCmd.Stdin = dumpOut

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("[MySQL] ❌ Failed to create output file: %v\n", err)
		return
	}
	defer outFile.Close()
	gzipCmd.Stdout = outFile

	// Start commands
	if err := dumpCmd.Start(); err != nil {
		fmt.Printf("[MySQL] ❌ mysqldump start error: %v\n", err)
		return
	}
	if err := gzipCmd.Start(); err != nil {
		fmt.Printf("[MySQL] ❌ gzip start error: %v\n", err)
		return
	}

	// Wait for completion
	if err := dumpCmd.Wait(); err != nil {
		fmt.Printf("[MySQL] ❌ mysqldump execution error: %v\n", err)
	}
	if err := gzipCmd.Wait(); err != nil {
		fmt.Printf("[MySQL] ❌ gzip execution error: %v\n", err)
	}

	fmt.Println("[MySQL] ✅ Backup completed successfully")
}
