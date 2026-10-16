package main

import (
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/internal/infra/server"
)

func main() {
	server.NewHTTPServer().Start(env.AppEnv.AppHost + ":" + env.AppEnv.AppPort)
}
