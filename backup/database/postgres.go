// 📄backup/postgres.go

package backup

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// postgresConfig holds PostgreSQL backup settings from environment variables.
type postgresConfig struct {
	dsn        string // full DSN, e.g. "postgres://user:pass@host:port/db"
	Host       string
	Port       string
	User       string
	Password   string
	Database   string // used when dsn is empty and DumpAll is false
	BackupDir  string
	UseDumpAll bool // when true, run pg_dumpall
}

// To back up entire cluster, set POSTGRES_DUMP_ALL=true.
func loadPostgresConfig() (*postgresConfig, error) {
	dsn := os.Getenv("POSTGRES_DSN")

	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	backupDir := os.Getenv("POSTGRES_BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/home/hyper-backup/postgres"
	}

	useDumpAll := os.Getenv("POSTGRES_DUMP_ALL") == "true"

	if dsn != "" {
		if _, err := url.Parse(dsn); err != nil {
			return nil, fmt.Errorf("invalid POSTGRES_DSN: %v", err)
		}
	} else {
		if host == "" || user == "" || pass == "" {
			return nil, fmt.Errorf("POSTGRES_HOST, POSTGRES_USER and POSTGRES_PASSWORD must be set")
		}
		if !useDumpAll && db == "" {
			return nil, fmt.Errorf("POSTGRES_DB must be set unless POSTGRES_DUMP_ALL=true or DSN provided")
		}
	}

	if port == "" {
		port = "5432"
	}

	return &postgresConfig{
		dsn:        dsn,
		Host:       host,
		Port:       port,
		User:       user,
		Password:   pass,
		Database:   db,
		BackupDir:  backupDir,
		UseDumpAll: useDumpAll,
	}, nil
}

func RunPostgres() {
	cfg, err := loadPostgresConfig()
	if err != nil {
		log.Printf("[PostgreSQL] ❌ Configuration error: %v\n", err)
		return
	}

	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		log.Printf("[PostgreSQL] ❌ Failed to create backup directory: %v\n", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	var filename string
	if cfg.UseDumpAll {
		filename = fmt.Sprintf("all_%s.sql.gz", timestamp)
	} else if cfg.dsn != "" {
		u, _ := url.Parse(cfg.dsn)
		dbname := filepath.Base(u.Path)
		if dbname == "" {
			dbname = "dsn"
		}
		filename = fmt.Sprintf("%s_%s.sql.gz", dbname, timestamp)
	} else {
		filename = fmt.Sprintf("%s_%s.sql.gz", cfg.Database, timestamp)
	}
	outputFile := filepath.Join(cfg.BackupDir, filename)

	log.Printf("[PostgreSQL] 🐘 Starting backup to %s\n", outputFile)

	if cfg.Password != "" {
		os.Setenv("PGPASSWORD", cfg.Password)
	}

	var cmd *exec.Cmd
	if cfg.UseDumpAll {
		if cfg.dsn != "" {
			cmd = exec.Command("pg_dumpall", "--dbname", cfg.dsn)
		} else {
			cmd = exec.Command(
				"pg_dumpall",
				"-h", cfg.Host,
				"-p", cfg.Port,
				"-U", cfg.User,
			)
		}
	} else {
		if cfg.dsn != "" {
			cmd = exec.Command("pg_dump", "--dbname", cfg.dsn)
		} else {
			cmd = exec.Command(
				"pg_dump",
				"-h", cfg.Host,
				"-p", cfg.Port,
				"-U", cfg.User,
				"-d", cfg.Database,
			)
		}
	}

	gzipCmd := exec.Command("gzip")
	dumpOut, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[PostgreSQL] ❌ Failed to pipe stdout: %v\n", err)
		return
	}
	gzipCmd.Stdin = dumpOut

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Printf("[PostgreSQL] ❌ Failed to create output file: %v\n", err)
		return
	}
	defer outFile.Close()
	gzipCmd.Stdout = outFile

	if err := cmd.Start(); err != nil {
		log.Printf("[PostgreSQL] ❌ Dump start error: %v\n", err)
		return
	}
	if err := gzipCmd.Start(); err != nil {
		log.Printf("[PostgreSQL] ❌ gzip start error: %v\n", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("[PostgreSQL] ❌ Dump execution error: %v\n", err)
	}
	if err := gzipCmd.Wait(); err != nil {
		log.Printf("[PostgreSQL] ❌ gzip execution error: %v\n", err)
	}

	log.Printf("[PostgreSQL] ✅ Backup completed successfully")
	log.Printf("\n")
}
