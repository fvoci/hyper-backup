// backup/mongo.go

package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// mongoConfig holds MongoDB backup settings loaded from environment variables.
type mongoConfig struct {
	URI       string // full MongoDB URI, e.g. "mongodb://user:pass@host:port/db"
	Host      string
	Port      string
	Database  string // specific database to dump; if empty dumps all
	BackupDir string
}

// loadMongoConfig reads and validates MongoDB backup settings.
// Supports full URI via MONGO_URI or individual vars.
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
		// validate URI
		u, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("invalid MONGO_URI: %v", err)
		}
		// extract database path if present
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

// RunMongo performs a dump of MongoDB (single DB or entire cluster) and compresses it to tar.gz.
func RunMongo() {
	cfg, err := loadMongoConfig()
	if err != nil {
		fmt.Printf("[MongoDB] ❌ Configuration error: %v\n", err)
		return
	}

	// Prepare timestamps and paths
	timestamp := time.Now().Format("20060102_150405")
	dumpDir := filepath.Join(cfg.BackupDir, fmt.Sprintf("dump_%s", timestamp))
	name := cfg.Database
	if name == "" {
		name = "all"
	}
	archive := fmt.Sprintf("%s_%s.tar.gz", name, timestamp)
	archivePath := filepath.Join(cfg.BackupDir, archive)

	// Ensure backup directory exists
	if err := os.MkdirAll(cfg.BackupDir, 0755); err != nil {
		fmt.Printf("[MongoDB] ❌ Failed to create backup directory: %v\n", err)
		return
	}

	fmt.Printf("[MongoDB] 🍃 Starting backup to %s\n", archivePath)

	// Build mongodump arguments
	var args []string
	if cfg.URI != "" {
		args = []string{"--uri=" + cfg.URI, "--out=" + dumpDir}
	} else {
		args = []string{"--host=" + cfg.Host, "--port=" + cfg.Port, "--out=" + dumpDir}
		if cfg.Database != "" {
			args = append(args, "--db="+cfg.Database)
		}
	}

	dumpCmd := exec.Command("mongodump", args...)
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr
	if err := dumpCmd.Run(); err != nil {
		fmt.Printf("[MongoDB] ❌ mongodump failed: %v\n", err)
		return
	}

	// Compress dump directory to tar.gz
	if err := createTarGz(archivePath, dumpDir); err != nil {
		fmt.Printf("[MongoDB] ❌ Compression failed: %v\n", err)
		return
	}

	// Remove temporary dump directory
	if err := os.RemoveAll(dumpDir); err != nil {
		fmt.Printf("[MongoDB] ⚠️ Failed to remove dump directory: %v\n", err)
	}

	fmt.Println("[MongoDB] ✅ Backup completed successfully")
}

// createTarGz creates a tar.gz archive from the source directory.
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
