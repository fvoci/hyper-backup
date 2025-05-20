// 📄 backup/traefik/traefik.go

package traefik

import (
	"log"
	"os"
)

func LogrotateAndNotify() {
	log.Printf("[Traefik] 🌀 Starting logrotate and notify process...")

	logFile := os.Getenv("TRAEFIK_LOG_FILE")
	if logFile == "" {
		log.Printf("[Traefik] ⚠️ TRAEFIK_LOG_FILE is not set")
		return
	}

	rotatedPath, copiedBytes, err := RotateAndBackup(logFile)
	if err != nil {
		log.Printf("[Traefik] ❌ Failed to rotate: %v", err)
		return
	}

	if copiedBytes == 0 {
		log.Printf("[Traefik] 💤 Log file empty, skipping rotation")
		return
	}

	log.Printf("[Traefik] 🔄 Copied %d bytes → %s", copiedBytes, rotatedPath)

	containerID, err := GetTraefikContainerID()
	if err != nil {
		log.Printf("[Traefik] ⚠️ No container found: %v", err)
	} else if err := SendUSR1(containerID); err != nil {
		log.Printf("[Traefik] ❌ Failed to send USR1: %v", err)
	} else {
		log.Printf("[Traefik] 📤 Rotated log: %s", rotatedPath)
		log.Printf("[Traefik] ✅ Logrotate and signal complete.")
	}
}
