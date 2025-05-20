package backup

import (
	db "github.com/fvoci/hyper-backup/backup/database"
	"github.com/fvoci/hyper-backup/backup/traefik"
	"github.com/fvoci/hyper-backup/utilities"
)

// RunCoreServices는 MySQL, PostgreSQL, MongoDB, Traefik 등 핵심 백업 서비스를 실행하고, 실행 중 발생한 오류를 반환합니다.
func RunCoreServices() error {
	utilities.LogDivider()
	utilities.Logger.Info("🔧 [Core Services]")

	services := []service{
		{
			Name:     "MySQL",
			EnvKeys:  []string{"MYSQL_HOST"},
			RunFunc:  db.RunMySQL,
			Optional: true,
		},
		{
			Name:     "PostgreSQL",
			EnvKeys:  []string{"POSTGRES_HOST"},
			RunFunc:  db.RunPostgres,
			Optional: true,
		},
		{
			Name:     "MongoDB",
			EnvKeys:  []string{"MONGO_HOST"},
			RunFunc:  db.RunMongo,
			Optional: true,
		},
		{
			Name:     "Traefik",
			EnvKeys:  []string{"TRAEFIK_LOG_FILE"},
			RunFunc:  traefik.LogrotateAndNotify,
			Optional: true,
		},
	}

	return runServices(services)
}
