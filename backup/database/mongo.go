package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fvoci/hyper-backup/utilities"
)

type mongoConfig struct {
	URI       string
	Host      string
	Port      string
	Database  string
	BackupDir string
}

// loadMongoConfig는 환경 변수에서 MongoDB 백업 설정을 로드하여 mongoConfig 구조체를 반환합니다.
// MONGO_URI가 설정된 경우 URI를 파싱하여 데이터베이스 이름을 추출하고, 그렇지 않으면 호스트와 포트 정보를 사용합니다.
// 필수 값이 누락되었거나 URI가 잘못된 경우 오류를 반환합니다.
func loadMongoConfig() (*mongoConfig, error) {
	uri := os.Getenv("MONGO_URI")
	host := os.Getenv("MONGO_HOST")
	port := os.Getenv("MONGO_PORT")
	database := os.Getenv("MONGO_DB")
	backupDir := os.Getenv("MONGO_BACKUP_DIR")
	if backupDir == "" {
		backupDir = "/home/hyper-backup/mongo"
	}
	if uri != "" {
		u, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("invalid MONGO_URI: %v", err)
		}
		database = strings.TrimPrefix(u.Path, "/")
	} else {
		if host == "" {
			return nil, fmt.Errorf("MONGO_HOST is required when MONGO_URI is not set")
		}
		if port == "" {
			port = "27017"
		}
	}
	return &mongoConfig{
		URI:       uri,
		Host:      host,
		Port:      port,
		Database:  database,
		BackupDir: backupDir,
	}, nil
}

// RunMongo는 MongoDB 데이터베이스를 덤프하고 tar.gz 아카이브로 압축하여 백업을 수행한다.
// 환경 변수에서 설정을 로드하고, 백업 디렉터리를 생성한 뒤, mongodump 명령을 실행하고, 결과를 압축한 후 임시 파일을 정리한다.
// 백업 과정에서 오류가 발생하면 해당 오류를 반환한다.
func RunMongo() error {
	cfg, err := loadMongoConfig()
	if err != nil {
		utilities.Logger.Errorf("[MongoDB] ❌ Configuration error: %v", err)
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	dumpDir := filepath.Join(cfg.BackupDir, fmt.Sprintf("dump_%s", timestamp))
	name := cfg.Database
	if name == "" {
		name = "all"
	}
	archive := fmt.Sprintf("%s_%s.tar.gz", name, timestamp)
	archivePath := filepath.Join(cfg.BackupDir, archive)

	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		utilities.Logger.Errorf("[MongoDB] ❌ Failed to create backup directory: %v", err)
		return err
	}

	utilities.Logger.Infof("[MongoDB] 🍃 Backing up database '%s' to: %s", name, archivePath)

	var out bytes.Buffer
	dumpCmd := exec.Command("mongodump", buildMongodumpArgs(cfg, dumpDir)...)
	dumpCmd.Stdout = &out
	dumpCmd.Stderr = &out

	if err := dumpCmd.Run(); err != nil {
		utilities.Logger.Errorf("[MongoDB] ❌ mongodump failed: %v", err)
		return err
	}

	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "done dumping") || strings.Contains(line, "writing") {
			utilities.Logger.Debugf("[MongoDB] 📄 %s", line)
		}
	}

	if err := createTarGz(archivePath, dumpDir); err != nil {
		utilities.Logger.Errorf("[MongoDB] ❌ Compression failed: %v", err)
		return err
	}

	if err := os.RemoveAll(dumpDir); err != nil {
		utilities.Logger.Warnf("[MongoDB] ⚠️ Failed to remove dump directory: %v", err)
	}

	utilities.Logger.Infof("[MongoDB] ✅ Backup of '%s' completed successfully", name)
	utilities.LogDivider()
	return nil
}

// buildMongodumpArgs는 주어진 MongoDB 설정과 덤프 디렉터리를 기반으로 mongodump 명령어 인자 목록을 생성합니다.
func buildMongodumpArgs(cfg *mongoConfig, dumpDir string) []string {
	if cfg.URI != "" {
		return []string{"--uri=" + cfg.URI, "--out=" + dumpDir}
	}
	args := []string{"--host=" + cfg.Host, "--port=" + cfg.Port, "--out=" + dumpDir}
	if cfg.Database != "" {
		args = append(args, "--db="+cfg.Database)
	}
	return args
}

func createTarGz(target, sourceDir string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
}
