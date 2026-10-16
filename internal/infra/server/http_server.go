package server

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/internal/middlewares"
	"github.com/mqqff/wasteco-be/pkg/log"
	"time"
)

type HTTPServerItf interface {
	Start(socket string)
	MountMiddlewares()
	GetApp() *fiber.App
}

type HTTPServer struct {
	app *fiber.App
}

func NewHTTPServer() HTTPServerItf {
	config := fiber.Config{
		AppName:       env.AppEnv.AppName,
		CaseSensitive: true,
		ServerHeader:  env.AppEnv.AppEnv,
		IdleTimeout:   10 * time.Second,
		JSONDecoder:   sonic.Unmarshal,
		JSONEncoder:   sonic.Marshal,
	}

	app := fiber.New(config)

	return &HTTPServer{app}
}

func (s *HTTPServer) GetApp() *fiber.App {
	return s.app
}

func (s *HTTPServer) Start(socket string) {
	err := s.app.Listen(socket)

	if err != nil {
		log.Fatal(log.LogInfo{
			"error": err.Error(),
		}, "[HTTPServer][Start] failed to start server")
	}
}

func (s *HTTPServer) MountMiddlewares() {
	s.app.Use(middlewares.Logger())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type", "Authorization", "X-API-KEY", "x-forwarded-for"},
	}))

	s.app.Use(helmet.New())
	s.app.Use(compress.New())
	s.app.Use(middlewares.RequireAPIKey())
}
