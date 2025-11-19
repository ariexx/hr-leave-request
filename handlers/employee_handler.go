package handlers

import (
	"hr-leave-request/dtos"
	"hr-leave-request/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type EmployeeHandler struct {
	service services.EmployeeService
}

func NewEmployeeHandler(service services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{service: service}
}

func (h *EmployeeHandler) CreateEmployee(c *fiber.Ctx) error {
	var req dtos.CreateEmployeeRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
	}

	employee, err := h.service.CreateEmployee(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create employee")
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "email already exists" {
			statusCode = fiber.StatusConflict
		}
		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Create Failed",
			Message: err.Error(),
		})
	}

	logrus.WithField("employee_id", employee.ID).Info("Employee created successfully")
	return c.Status(fiber.StatusCreated).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Employee created successfully",
		Data:    employee,
	})
}

func (h *EmployeeHandler) GetEmployeeByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid employee ID",
		})
	}

	employee, err := h.service.GetEmployeeByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Failed to get employee")
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "employee not found" {
			statusCode = fiber.StatusNotFound
		}
		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Get Failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Employee retrieved successfully",
		Data:    employee,
	})
}

func (h *EmployeeHandler) GetEmployees(c *fiber.Ctx) error {
	var req dtos.GetEmployeesRequest

	// Parse query parameters
	if err := c.QueryParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse query parameters")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
	}

	employees, err := h.service.GetEmployees(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get employees")
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error:   "Get Failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Employees retrieved successfully",
		Data:    employees,
	})
}
