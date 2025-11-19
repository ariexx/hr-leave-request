package handlers

import (
	"hr-leave-request/dtos"
	"hr-leave-request/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService services.AuthService, validator *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator,
	}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dtos.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dtos.ErrorResponse{
			Error:   "unauthorized",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dtos.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid request body",
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: err.Error(),
		})
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}
