package backup

import (
	"os"

	utiles "github.com/fvoci/hyper-backup/utilities"
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
			utiles.Logger.Infof("[%s] ▶️ Starting backup...", svc.Name)
			safeRun(svc.Name, svc.RunFunc)
			executed++
		} else if !svc.Optional {
			utiles.Logger.Infof("[%s] ⚠️ Required but not configured; skipping", svc.Name)
		}
	}
	if executed == 0 {
		utiles.Logger.Warn("🤷 No services matched conditions")
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

// safeRun executes a backup task and recovers from panics or errors
func safeRun(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			utiles.Logger.Errorf("[%s] ❌ Panic during backup: %v", name, r)
		}
	}()
	fn()
}
