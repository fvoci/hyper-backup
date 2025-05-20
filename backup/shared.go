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

// runServices는 주어진 서비스 목록 중 환경 변수 조건을 만족하는 백업 서비스를 실행하고, 발생한 모든 오류를 결합하여 반환합니다.
// 필수 서비스가 환경 변수로 구성되지 않은 경우에도 오류로 처리됩니다. 
// 실행된 서비스가 없고 오류도 없는 경우 경고 로그를 남깁니다.
//
// 반환값:
//   - 실행 중 발생한 모든 오류를 결합한 error. 오류가 없으면 nil을 반환합니다.
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
			msg := fmt.Errorf("required service not configured")
			utilities.Logger.Errorf("[%s] ❌ %v", svc.Name, msg)
			errs = append(errs, fmt.Errorf("%s: %w", svc.Name, msg))
		}
	}

	if executed == 0 && len(errs) == 0 {
		utilities.Logger.Warn("🤷 No services matched conditions")
	}

	return errors.Join(errs...)
}

// shouldRun은 주어진 모든 환경 변수 키가 설정되어 있는지 확인하여, 모두 설정되어 있으면 true를 반환합니다.
func shouldRun(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return false
		}
	}
	return true
}

// safeRunWithError는 주어진 함수 실행 중 발생하는 패닉을 복구하여 에러로 반환합니다.
// 패닉이 발생하면 스택 트레이스와 함께 에러로 변환하며, 그렇지 않으면 함수의 반환 에러를 그대로 반환합니다.
func safeRunWithError(name string, fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			utilities.Logger.Errorf("[%s] 💥 panic recovered: %v\n%s", name, r, stack)
			err = fmt.Errorf("panic in [%s]: %v\n%s", name, r, string(stack))
		}
	}()
	return fn()
}
