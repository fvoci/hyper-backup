// 📄 backup/service.go

package backup

import (
	"log"

	db "github.com/fvoci/hyper-backup/backup/database"
	"github.com/fvoci/hyper-backup/backup/traefik"
)

func RunCoreServices() {
	log.Printf("\n")
	log.Printf("🔧 [Core Services]")
	services := []service{
		{"MySQL", []string{"MYSQL_HOST"}, db.RunMySQL, true},
		{"PostgreSQL", []string{"POSTGRES_HOST"}, db.RunPostgres, true},
		{"MongoDB", []string{"MONGO_HOST"}, db.RunMongo, true},
		{"Traefik", []string{"TRAEFIK_LOG_FILE"}, traefik.LogrotateAndNotify, true},
	}
	runServices(services)
}
