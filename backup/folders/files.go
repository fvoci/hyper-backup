package folders

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/fvoci/hyper-backup/utilities"
)

// RunFileBackup compresses directories defined by PACK_UP_HYPER_BACKUP_* env vars.
// Returns a list of created archive file paths.
func RunFileBackup() []string {
	baseDir := "/home/hyper-backup/files"
	_ = os.MkdirAll(baseDir, 0755)

	var created []string

	for i := 1; ; i++ {
		envKey := fmt.Sprintf("PACK_UP_HYPER_BACKUP_%d", i)
		src := os.Getenv(envKey)
		if src == "" {
			break
		}

		if fi, err := os.Stat(src); err != nil || !fi.IsDir() {
			utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: not a valid directory", src)
			continue
		}

		timestamp := time.Now().Format("20060102_150405")
		name := strings.ReplaceAll(filepath.Base(src), " ", "_")

		method := os.Getenv("FILE_BACKUP_COMPRESSION")
		if method == "" {
			method = "zstd"
		}

		var outPath string
		var err error

		switch strings.ToLower(method) {
		case "gzip":
			outPath = filepath.Join(baseDir, fmt.Sprintf("%s_%s.tar.gz", name, timestamp))
			err = compressToTarGz(src, outPath)
		case "zstd":
			outPath = filepath.Join(baseDir, fmt.Sprintf("%s_%s.tar.zst", name, timestamp))
			err = compressToTarZst(src, outPath)
		default:
			utilities.Logger.Errorf("[Files] ❌ Unknown compression method: %s", method)
			continue
		}

		if err != nil {
			utilities.Logger.Errorf("[Files] ❌ Failed to compress %s: %v", src, err)
			continue
		}

		utilities.Logger.Infof("[Files] 📦 Packed %s → %s", src, outPath)
		created = append(created, outPath)
	}

	if len(created) == 0 {
		utilities.Logger.Info("[Files] 🤷 No folders were packed")
	}
	utilities.LogDivider()
	return created
}

// compressToTarGz compresses a directory into .tar.gz
func compressToTarGz(srcDir, outFile string) error {
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return walkAndWriteTar(srcDir, tw)
}

// compressToTarZst compresses a directory into .tar.zst
func compressToTarZst(srcDir, outFile string) error {
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	zw, err := zstd.NewWriter(out)
	if err != nil {
		return err
	}
	defer zw.Close()

	tw := tar.NewWriter(zw)
	defer tw.Close()

	return walkAndWriteTar(srcDir, tw)
}

func walkAndWriteTar(srcDir string, tw *tar.Writer) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: stat error: %v", path, err)
			return nil
		}
		relPath, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: rel path error: %v", path, err)
			return nil
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: header error: %v", path, err)
			return nil
		}
		hdr.Name = relPath
		if err := tw.WriteHeader(hdr); err != nil {
			utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: write header error: %v", path, err)
			return nil
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: open error: %v", path, err)
				return nil
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				utilities.Logger.Warnf("[Files] ⚠️ Skipping %s: copy error: %v", path, err)
			}
		}
		return nil
	})
}
