package server

import (
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/pkg/log"
)

type HTTPServerItf interface {
	Start(socket string)
	GetApp() *fiber.App
}

type HTTPServer struct {
	app *fiber.App
}

func NewHTTPServer() HTTPServerItf {
	app := fiber.New(fiber.Config{
		AppName: env.AppEnv.AppName, CaseSensitive: true,
		ServerHeader: env.AppEnv.AppEnv, IdleTimeout: 10 * time.Second,
		JSONDecoder: sonic.Unmarshal, JSONEncoder: sonic.Marshal,
	})
	return &HTTPServer{app: app}
}

func (s *HTTPServer) GetApp() *fiber.App {
	return s.app
}

func (s *HTTPServer) Start(socket string) {
	if err := s.app.Listen(socket); err != nil {
		log.Fatal(log.LogInfo{"error": err.Error()}, "[HTTPServer][Start] failed to start server")
	}
}
