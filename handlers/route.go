package handlers

import (
	"hr-leave-request/config"
	"hr-leave-request/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func SetupRoutes(app *fiber.App, employeeHandler *EmployeeHandler, authHandler *AuthHandler, leaveRequestHandler *LeaveRequestHandler, cfg *config.ApplicationConfig) {
	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "HR Leave Request API is running",
		})
	})

	// API v1 routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.Post("/login", authHandler.Login)
		auth.Post("/register", authHandler.Register)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware(cfg))

	// Employee routes (protected)
	employees := protected.Group("/employees")
	{
		employees.Post("/", employeeHandler.CreateEmployee)
		employees.Get("/", employeeHandler.GetEmployees)
		employees.Get("/:id", employeeHandler.GetEmployeeByID)
	}

	// Leave request routes (protected)
	leaveRequests := protected.Group("/leave-requests")
	{
		leaveRequests.Post("/", leaveRequestHandler.CreateLeaveRequest)
		leaveRequests.Get("/", leaveRequestHandler.GetLeaveRequests)
		leaveRequests.Get("/:id", leaveRequestHandler.GetLeaveRequestByID)
		leaveRequests.Put("/:id", leaveRequestHandler.UpdateLeaveRequest)
		leaveRequests.Delete("/:id", leaveRequestHandler.DeleteLeaveRequest)
	}
}
