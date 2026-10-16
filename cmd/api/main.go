package main

import (
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/internal/infra/mysql"
	"github.com/mqqff/wasteco-be/internal/infra/server"
)

func main() {
	db, err := mysql.NewConn()
	if err != nil {
		return
	}

	httpServer := server.NewHTTPServer()
	httpServer.MountMiddlewares()
	httpServer.MountRoutes(db)
	httpServer.Start(env.AppEnv.AppHost + ":" + env.AppEnv.AppPort)
}
