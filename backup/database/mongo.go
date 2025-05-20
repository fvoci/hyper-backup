// 📄 backup/database/mongo.go

package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type mongoConfig struct {
	URI       string
	Host      string
	Port      string
	Database  string
	BackupDir string
}

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

func RunMongo() {
	cfg, err := loadMongoConfig()
	if err != nil {
		log.Printf("[MongoDB] ❌ Configuration error: %v", err)
		return
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
		log.Printf("[MongoDB] ❌ Failed to create backup directory: %v", err)
		return
	}

	log.Printf("[MongoDB] 🍃 Starting backup to %s", archivePath)

	var out bytes.Buffer
	dumpCmd := exec.Command("mongodump", buildMongodumpArgs(cfg, dumpDir)...)
	dumpCmd.Stdout = &out
	dumpCmd.Stderr = &out

	if err := dumpCmd.Run(); err != nil {
		log.Printf("[MongoDB] ❌ mongodump failed: %v", err)
	}

	for _, line := range strings.Split(out.String(), "\n") {
		if strings.TrimSpace(line) != "" {
			log.Printf("[MongoDB] 📄 %s", line)
		}
	}

	if err := createTarGz(archivePath, dumpDir); err != nil {
		log.Printf("[MongoDB] ❌ Compression failed: %v", err)
		return
	}

	if err := os.RemoveAll(dumpDir); err != nil {
		log.Printf("[MongoDB] ⚠️ Failed to remove dump directory: %v", err)
	}

	log.Printf("[MongoDB] ✅ Backup completed successfully")
	log.Printf("\n")
}

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
