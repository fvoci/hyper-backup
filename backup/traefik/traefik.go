package traefik

import (
	"os"

	utiles "github.com/fvoci/hyper-backup/utilities"
)

func LogrotateAndNotify() {
	utiles.Logger.Info("[Traefik] 🌀 Starting logrotate and notify process...")

	logFile := os.Getenv("TRAEFIK_LOG_FILE")
	if logFile == "" {
		utiles.Logger.Warn("[Traefik] ⚠️ TRAEFIK_LOG_FILE is not set")
		return
	}

	rotatedPath, copiedBytes, err := RotateAndBackup(logFile)
	if err != nil {
		utiles.Logger.Errorf("[Traefik] ❌ Failed to rotate: %v", err)
		return
	}

	if copiedBytes == 0 {
		utiles.Logger.Info("[Traefik] 💤 Log file empty, skipping rotation")
		return
	}

	utiles.Logger.Infof("[Traefik] 🔄 Copied %d bytes → %s", copiedBytes, rotatedPath)

	containerID, err := GetTraefikContainerID()
	if err != nil {
		utiles.Logger.Warnf("[Traefik] ⚠️ No container found: %v", err)
	} else if err := SendUSR1(containerID); err != nil {
		utiles.Logger.Errorf("[Traefik] ❌ Failed to send USR1: %v", err)
	} else {
		utiles.Logger.Infof("[Traefik] 📤 Rotated log: %s", rotatedPath)
		utiles.Logger.Info("[Traefik] ✅ Logrotate and signal complete.")
	}
}
