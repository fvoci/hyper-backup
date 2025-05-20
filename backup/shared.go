package backup

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

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
			err := safeRunWithError(svc.Name, svc.RunFunc)
			if err != nil {
				utilities.Logger.Errorf("[%s] ❌ Backup failed: %v", svc.Name, err)
				errs = append(errs, fmt.Errorf("%s: %w", svc.Name, err))
			}
			executed++
		} else if !svc.Optional {
			utilities.Logger.Warnf("[%s] ⚠️ Required but not configured; skipping", svc.Name)
		}
	}

	if executed == 0 {
		utilities.Logger.Warn("🤷 No services matched conditions")
	}

	if len(errs) > 0 {
		return fmt.Errorf("some services failed: %w", joinErrors(errs))
	}

	return nil
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

// safeRunWithError catches panic and wraps it into error
func safeRunWithError(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			utilities.Logger.Errorf("[%s] 💥 panic recovered: %v\n%s", name, r, string(stack))
			err = fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack())
		}
	}()
	return fn()
}

// joinErrors formats multiple errors into one
func joinErrors(errs []error) error {
	if len(errs) == 1 {
		return errs[0]
	}
	var b strings.Builder
	for _, err := range errs {
		fmt.Fprintf(&b, "\n- %v", err)
	}
	return fmt.Errorf("multiple errors:%s", b.String())
}
