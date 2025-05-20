package traefik

import (
	"fmt"
	"os"

	"github.com/fvoci/hyper-backup/utilities"
)

// LogrotateAndNotify는 Traefik 로그 파일을 회전(로테이션)하고, Traefik 컨테이너에 USR1 시그널을 전송하여 로그 파일 변경을 알립니다.
//
// TRAEFIK_LOG_FILE 환경 변수가 설정되어 있지 않으면 오류를 반환하며, 로그 파일이 비어 있으면 아무 작업도 수행하지 않습니다.
// 로그 회전 또는 시그널 전송 중 오류가 발생하면 해당 오류를 반환합니다.
//
// 성공적으로 완료되면 nil을 반환합니다.
func LogrotateAndNotify() error {
	utilities.Logger.Info("[Traefik] 🌀 Starting logrotate and notify process...")

	logFile := os.Getenv("TRAEFIK_LOG_FILE")
	if logFile == "" {
		utilities.Logger.Warn("[Traefik] ⚠️ TRAEFIK_LOG_FILE is not set")
		return fmt.Errorf("TRAEFIK_LOG_FILE is not set")
	}

	rotatedPath, copiedBytes, err := RotateAndBackup(logFile)
	if err != nil {
		utilities.Logger.Errorf("[Traefik] ❌ Failed to rotate: %v", err)
		return err
	}

	if copiedBytes == 0 {
		utilities.Logger.Info("[Traefik] 💤 Log file empty, skipping rotation")
		return nil
	}

	utilities.Logger.Infof("[Traefik] 🔄 Copied %d bytes → %s", copiedBytes, rotatedPath)

	containerID, err := GetTraefikContainerID()
	if err != nil {
		utilities.Logger.Warnf("[Traefik] ⚠️ No container found: %v", err)
	} else if err := SendUSR1(containerID); err != nil {
		utilities.Logger.Errorf("[Traefik] ❌ Failed to send USR1: %v", err)
		return err
	} else {
		utilities.Logger.Infof("[Traefik] 📤 Rotated log: %s", rotatedPath)
		utilities.Logger.Info("[Traefik] ✅ Logrotate and signal complete.")
		utilities.LogDivider()
	}

	return nil
}
