package backup

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fvoci/hyper-backup/utilities"
)

// postgresConfig holds PostgreSQL backup settings from environment variables.
type postgresConfig struct {
	dsn        string // full DSN, e.g. "postgres://user:pass@host:port/db"
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	BackupDir  string
	UseDumpAll bool
}

// loadPostgresConfig는 환경 변수에서 PostgreSQL 백업 설정을 읽어와 검증한 후 postgresConfig 구조체를 반환합니다.
// DSN이 제공되면 이를 우선 사용하며, 그렇지 않은 경우 개별 연결 정보가 모두 설정되어 있어야 합니다.
// 필수 값이 누락되었거나 DSN이 잘못된 경우 오류를 반환합니다.
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

// RunPostgres는 환경 변수로 구성된 설정을 사용하여 PostgreSQL 데이터베이스의 백업을 수행하고, 압축된 백업 파일을 지정된 디렉터리에 저장합니다.
// 백업 과정에서 오류가 발생하면 해당 오류를 반환합니다.
func RunPostgres() error {
	cfg, err := loadPostgresConfig()
	if err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Configuration error: %v", err)
		return err
	}

	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Failed to create backup directory: %v", err)
		return err
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

	utilities.Logger.Infof("[PostgreSQL] 🐘 Starting backup to %s", outputFile)

	if cfg.Password != "" {
		os.Setenv("PGPASSWORD", cfg.Password)
	}

	var cmd *exec.Cmd
	if cfg.UseDumpAll {
		if cfg.dsn != "" {
			cmd = exec.Command("pg_dumpall", "--dbname", cfg.dsn)
		} else {
			cmd = exec.Command("pg_dumpall", "-h", cfg.Host, "-p", cfg.Port, "-U", cfg.User)
		}
	} else {
		if cfg.dsn != "" {
			cmd = exec.Command("pg_dump", "--dbname", cfg.dsn)
		} else {
			cmd = exec.Command("pg_dump", "-h", cfg.Host, "-p", cfg.Port, "-U", cfg.User, "-d", cfg.Database)
		}
	}

	gzipCmd := exec.Command("gzip")
	dumpOut, err := cmd.StdoutPipe()
	if err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Failed to pipe stdout: %v", err)
		return err
	}
	gzipCmd.Stdin = dumpOut

	outFile, err := os.Create(outputFile)
	if err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Failed to create output file: %v", err)
		return err
	}
	defer outFile.Close()
	gzipCmd.Stdout = outFile

	if err := cmd.Start(); err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Dump start error: %v", err)
		return err
	}
	if err := gzipCmd.Start(); err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ gzip start error: %v", err)
		return err
	}

	if err := cmd.Wait(); err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ Dump execution error: %v", err)
		return err
	}
	if err := gzipCmd.Wait(); err != nil {
		utilities.Logger.Errorf("[PostgreSQL] ❌ gzip execution error: %v", err)
		return err
	}

	utilities.Logger.Info("[PostgreSQL] ✅ Backup completed successfully")
	utilities.LogDivider()
	return nil
}
