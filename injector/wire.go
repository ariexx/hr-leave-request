//go:build wireinject
// +build wireinject

package injector

import (
	"hr-leave-request/config"
	"hr-leave-request/handlers"
	"hr-leave-request/repositories"
	"hr-leave-request/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
)

func InitializeApp() (*fiber.App, error) {
	wire.Build(
		config.LoadConfig,
		config.NewDatabase,
		repositories.NewEmployeeRepository,
		services.NewEmployeeService,
		handlers.NewEmployeeHandler,
		NewFiberApp,
	)
	return nil, nil
}

func NewFiberApp(employeeHandler *handlers.EmployeeHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "HR Leave Request API",
	})

	handlers.SetupRoutes(app, employeeHandler)

	return app
}
