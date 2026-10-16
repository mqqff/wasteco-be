package main

import (
	"github.com/mqqff/wasteco-be/internal/domain/entity"
	"github.com/mqqff/wasteco-be/internal/infra/mysql"
	"github.com/mqqff/wasteco-be/pkg/log"
)

func main() {
	db, err := mysql.NewConn()
	if err != nil {
		return
	}

	err = db.Migrator().CreateTable(&entity.User{})
	if err != nil {
		log.Error(log.LogInfo{
			"error": err.Error(),
		}, "[Migration][main] failed to create tables")
		return
	}

	log.Info(nil, "[Migration][main] migration success")
}
