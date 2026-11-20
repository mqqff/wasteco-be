package server

import (
	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	AuthHandler "github.com/mqqff/wasteco-be/internal/app/auth/interface/rest"
	AuthUsecase "github.com/mqqff/wasteco-be/internal/app/auth/usecase"
	DonationHandler "github.com/mqqff/wasteco-be/internal/app/donation/interface/rest"
	DonationRepo "github.com/mqqff/wasteco-be/internal/app/donation/repository"
	DonationUsecase "github.com/mqqff/wasteco-be/internal/app/donation/usecase"
	ImpactHandler "github.com/mqqff/wasteco-be/internal/app/impact/interface/rest"
	ImpactRepo "github.com/mqqff/wasteco-be/internal/app/impact/repository"
	ImpactUsecase "github.com/mqqff/wasteco-be/internal/app/impact/usecase"
	PantryHandler "github.com/mqqff/wasteco-be/internal/app/pantry/interface/rest"
	PantryRepo "github.com/mqqff/wasteco-be/internal/app/pantry/repository"
	PantryUsecase "github.com/mqqff/wasteco-be/internal/app/pantry/usecase"
	RecipeHandler "github.com/mqqff/wasteco-be/internal/app/recipe/interface/rest"
	RecipeUsecase "github.com/mqqff/wasteco-be/internal/app/recipe/usecase"
	UserHandler "github.com/mqqff/wasteco-be/internal/app/user/interface/rest"
	UserRepo "github.com/mqqff/wasteco-be/internal/app/user/repository"
	UserUsecase "github.com/mqqff/wasteco-be/internal/app/user/usecase"
	"github.com/mqqff/wasteco-be/internal/domain/env"
	"github.com/mqqff/wasteco-be/internal/infra/gemini"
	"github.com/mqqff/wasteco-be/internal/middlewares"
	"github.com/mqqff/wasteco-be/pkg/jwt"
	"github.com/mqqff/wasteco-be/pkg/log"
	"gorm.io/gorm"
	"time"
)

type HTTPServerItf interface {
	Start(socket string)
	MountMiddlewares()
	MountRoutes(db *gorm.DB)
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

func (s *HTTPServer) MountRoutes(db *gorm.DB) {
	app := s.app

	api := app.Group("/api")
	v1 := api.Group("/v1")

	jwt := jwt.JWT
	validator := validator.New()
	geminiAI := gemini.Gemini

	middleware := middlewares.NewMiddleware(jwt)

	userRepo := UserRepo.NewUserRepo(db)
	pantryRepo := PantryRepo.NewPantryRepo(db)
	impactRepo := ImpactRepo.NewImpactRepo(db)
	donationRepo := DonationRepo.NewDonationRepo(db)

	authUsecase := AuthUsecase.NewAuthUsecase(userRepo, jwt)
	userUsecase := UserUsecase.NewUserUsecase(userRepo)
	pantryUsecase := PantryUsecase.NewPantryUsecase(pantryRepo, geminiAI)
	recipeUsecase := RecipeUsecase.NewRecipeUsecase(pantryRepo, impactRepo, geminiAI)
	donationUsecase := DonationUsecase.NewDonationUsecase(donationRepo, pantryRepo, userRepo)
	impactUsecase := ImpactUsecase.NewImpactUsecase(impactRepo)

	AuthHandler.NewAuthHandler(v1, middleware, authUsecase, validator)
	UserHandler.NewUserHandler(v1, middleware, userUsecase, validator)
	PantryHandler.NewPantryHandler(v1, middleware, pantryUsecase, validator)
	RecipeHandler.NewRecipeHandler(v1, middleware, recipeUsecase, validator)
	DonationHandler.NewDonationHandler(v1, middleware, donationUsecase, validator)
	ImpactHandler.NewImpactHandler(v1, middleware, impactUsecase)
}
