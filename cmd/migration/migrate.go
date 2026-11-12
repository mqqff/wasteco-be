package main

import (
	"flag"

	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/mysql"
	"github.com/mqqff/wasteco-be/pkg/log"
)

func main() {
	fresh := flag.Bool("fresh", false, "drop all tables before migrating")
	flag.Parse()

	db, err := mysql.NewConn()
	if err != nil {
		return
	}

	if *fresh {
		log.Info(nil, "[Migration][main] dropping existing tables...")
		err = db.Migrator().DropTable(
			&entity.ImpactActivity{},
			&entity.Donation{},
			&entity.PantryIngredient{},
			&entity.User{},
		)
		if err != nil {
			log.Error(log.LogInfo{
				"error": err.Error(),
			}, "[Migration][main] failed to drop tables")
			return
		}
	}

	log.Info(nil, "[Migration][main] migrating tables...")
	err = db.AutoMigrate(
		&entity.User{},
		&entity.PantryIngredient{},
		&entity.Donation{},
		&entity.ImpactActivity{},
	)
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Migration][main] failed to auto-migrate tables")
		return
	}

	log.Info(nil, "[Migration][main] migration success")
}
