//go:build wireinject
// +build wireinject

package injector

import (
	"hr-leave-request/config"
	"hr-leave-request/handlers"
	"hr-leave-request/repositories"
	"hr-leave-request/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

func InitializeApp() (*fiber.App, error) {
	wire.Build(
		config.LoadConfig,
		config.NewDatabase,
		NewValidator,
		repositories.NewEmployeeRepository,
		services.NewEmployeeService,
		services.NewAuthService,
		handlers.NewEmployeeHandler,
		handlers.NewAuthHandler,
		NewFiberApp,
	)
	return nil, nil
}

func NewFiberApp(
	employeeHandler *handlers.EmployeeHandler,
	authHandler *handlers.AuthHandler,
	cfg *config.ApplicationConfig,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "HR Leave Request API",
	})

	handlers.SetupRoutes(app, employeeHandler, authHandler, cfg)

	return app
}

func NewValidator() *validator.Validate {
	return validator.New()
}
