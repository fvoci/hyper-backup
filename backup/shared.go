// 📄 backup/shared.go

package backup

import (
	"log"
	"os"
)

type service struct {
	Name     string
	EnvKeys  []string
	RunFunc  func()
	Optional bool
}

func runServices(services []service) {
	executed := 0
	for _, svc := range services {
		if shouldRun(svc.EnvKeys...) {
			log.Printf("[%s] ▶️ Starting backup...", svc.Name)
			svc.RunFunc()
			executed++
		} else if !svc.Optional {
			log.Printf("[%s] ⚠️ Required but not configured; skipping", svc.Name)
		}
	}
	if executed == 0 {
		log.Printf("🤷 No services matched conditions")
	}
}

// shouldRun returns true if all listed environment variables are set
func shouldRun(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return false
		}
	}
	return true
}
