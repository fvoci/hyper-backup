package backup

import (
	db "github.com/fvoci/hyper-backup/backup/database"
	"github.com/fvoci/hyper-backup/backup/traefik"
	utiles "github.com/fvoci/hyper-backup/utilities"
)

func RunCoreServices() {
	utiles.LogDivider()
	utiles.Logger.Info("🔧 [Core Services]")

	services := []service{
		{"MySQL", []string{"MYSQL_HOST"}, db.RunMySQL, true},
		{"PostgreSQL", []string{"POSTGRES_HOST"}, db.RunPostgres, true},
		{"MongoDB", []string{"MONGO_HOST"}, db.RunMongo, true},
		{"Traefik", []string{"TRAEFIK_LOG_FILE"}, traefik.LogrotateAndNotify, true},
	}
	runServices(services)
}
