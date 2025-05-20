package backup

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fvoci/hyper-backup/utilities"
)

type service struct {
	Name     string
	EnvKeys  []string
	RunFunc  func() error
	Optional bool
}

func runServices(services []service) error {
	var errs []error
	executed := 0

	for _, svc := range services {
		if shouldRun(svc.EnvKeys...) {
			utilities.Logger.Infof("[%s] ▶️ Starting backup...", svc.Name)
			if err := safeRunWithError(svc.Name, svc.RunFunc); err != nil {
				utilities.Logger.Errorf("[%s] ❌ Backup failed: %v", svc.Name, err)
				errs = append(errs, fmt.Errorf("%s: %w", svc.Name, err))
			}
			executed++
		} else if !svc.Optional {
			utilities.Logger.Warnf("[%s] ⚠️ Required but not configured; skipping", svc.Name)
		}
	}

	if executed == 0 && len(errs) == 0 {
		utilities.Logger.Warn("🤷 No services matched conditions")
	}

	return errors.Join(errs...)
}

func shouldRun(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return false
		}
	}
	return true
}

func safeRunWithError(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			utilities.Logger.Errorf("[%s] 💥 panic recovered: %v\n%s", name, r, stack)
			err = fmt.Errorf("panic in [%s]: %v\n%s", name, r, stack)
		}
	}()
	return fn()
}
