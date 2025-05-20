package traefik

import (
	"fmt"
	"os"

	"github.com/fvoci/hyper-backup/utilities"
)

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
	}

	return nil
}
