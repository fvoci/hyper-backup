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
// RunFileBackup는 환경 변수로 지정된 여러 디렉터리를 압축하여 아카이브 파일로 백업하고, 생성된 아카이브 파일 경로 목록을 반환합니다.
//
// 환경 변수 PACK_UP_HYPER_BACKUP_1, PACK_UP_HYPER_BACKUP_2 등으로 지정된 각 디렉터리를 순차적으로 읽어, 지정된 압축 방식(FILE_BACKUP_COMPRESSION, 기본값 "zstd")에 따라 압축합니다. 유효하지 않은 디렉터리나 알 수 없는 압축 방식은 건너뜁니다.
//
// 반환값:
//   생성된 아카이브 파일의 전체 경로 목록을 반환합니다.
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

// 개별 파일 처리 중 오류가 발생해도 전체 작업은 계속 진행됩니다.
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
