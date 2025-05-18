// backup/run_all.go

package backup

import (
	"fmt"
	"os"
)

// service represents a backup service with its environment condition, execution function, and optionality.
type service struct {
	Name     string   // Display name of the service
	EnvKeys  []string // Required environment variables for this service
	RunFunc  func()   // Function to execute the backup
	Optional bool     // Whether the service is optional
}

// RunAll triggers all configured backup services based on environment variables.
// Required services will warn if missing; optional services will be skipped silently.
func RunAll() {
	fmt.Println("[HyperBackup] 🛠️ Invoking backup services")

	services := []service{
		{"MySQL", []string{"MYSQL_HOST"}, RunMySQL, false},
		{"PostgreSQL", []string{"POSTGRES_HOST"}, RunPostgres, false},
		{"MongoDB", []string{"MONGO_HOST"}, RunMongo, false},
		{"Rclone", []string{"RCLONE_REMOTE", "RCLONE_PATH"}, RunRclone, true},
		{"Rsync", []string{"RSYNC_SRC", "RSYNC_DEST"}, RunRsync, true},
	}

	for _, svc := range services {
		if shouldRun(svc.EnvKeys...) {
			fmt.Printf("[%s] Conditions met; executing %s backup\n", svc.Name, svc.Name)
			svc.RunFunc()
		} else if !svc.Optional {
			fmt.Printf("[%s] ⚠️ Required backup not configured; skipping\n", svc.Name)
		}
		// Optional services are skipped silently when not configured
	}
}

// shouldRun checks if all required environment variables are set.
func shouldRun(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) == "" {
			return false
		}
	}
	return true
}
