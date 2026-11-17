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
			&entity.DonationRequestItem{},
			&entity.DonationRequest{},
			"donations",
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
		&entity.DonationRequest{},
		&entity.DonationRequestItem{},
		&entity.ImpactActivity{},
	)
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Migration][main] failed to auto-migrate tables")
		return
	}

	obsoleteColumns := []struct {
		model  any
		column string
	}{
		{&entity.PantryIngredient{}, "estimated_price"},
		{&entity.ImpactActivity{}, "money_saved"},
		{&entity.DonationRequestItem{}, "unit_price"},
	}
	for _, c := range obsoleteColumns {
		if !db.Migrator().HasColumn(c.model, c.column) {
			continue
		}
		if err := db.Migrator().DropColumn(c.model, c.column); err != nil {
			log.Error(log.LogInfo{
				"error":  err.Error(),
				"column": c.column,
			}, "[Migration][main] failed to drop obsolete column")
			return
		}
	}

	log.Info(nil, "[Migration][main] migration success")
}
