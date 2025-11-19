package handlers

import (
	"hr-leave-request/dtos"
	"hr-leave-request/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type LeaveRequestHandler struct {
	service services.LeaveRequestService
}

func NewLeaveRequestHandler(service services.LeaveRequestService) *LeaveRequestHandler {
	return &LeaveRequestHandler{service: service}
}

func (h *LeaveRequestHandler) CreateLeaveRequest(c *fiber.Ctx) error {
	var req dtos.CreateLeaveRequestRequest

	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
	}

	// Get user ID from JWT middleware
	userID := c.Locals("user_id").(uint)

	leaveRequest, err := h.service.CreateLeaveRequest(userID, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create leave request")
		statusCode := fiber.StatusInternalServerError
		message := err.Error()

		switch message {
		case "employee not found":
			statusCode = fiber.StatusNotFound
		case "start date cannot be after end date", "cannot create leave request for past dates", "overlapping approved leave request exists for this date range":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Create Failed",
			Message: message,
		})
	}

	logrus.WithField("leave_request_id", leaveRequest.ID).Info("Leave request created successfully")
	return c.Status(fiber.StatusCreated).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request created successfully",
		Data:    leaveRequest,
	})
}

func (h *LeaveRequestHandler) GetLeaveRequestByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid leave request ID",
		})
	}

	leaveRequest, err := h.service.GetLeaveRequestByID(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Failed to get leave request")
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "leave request not found" {
			statusCode = fiber.StatusNotFound
		}
		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Get Failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request retrieved successfully",
		Data:    leaveRequest,
	})
}

func (h *LeaveRequestHandler) GetLeaveRequests(c *fiber.Ctx) error {
	var req dtos.GetLeaveRequestsRequest

	// Parse query parameters
	if err := c.QueryParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse query parameters")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid query parameters",
			Details: err.Error(),
		})
	}

	leaveRequests, err := h.service.GetLeaveRequests(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to get leave requests")
		return c.Status(fiber.StatusInternalServerError).JSON(dtos.ErrorResponse{
			Error:   "Get Failed",
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave requests retrieved successfully",
		Data:    leaveRequests,
	})
}

func (h *LeaveRequestHandler) UpdateLeaveRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid leave request ID",
		})
	}

	var req dtos.UpdateLeaveRequestRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse request body")
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
			Details: err.Error(),
		})
	}

	// Get user info from JWT middleware
	userID := c.Locals("user_id").(uint)
	var userRole string
	if role := c.Locals("role"); role != nil {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	}

	leaveRequest, err := h.service.UpdateLeaveRequest(uint(id), userID, userRole, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update leave request")
		statusCode := fiber.StatusInternalServerError
		message := err.Error()

		switch message {
		case "leave request not found":
			statusCode = fiber.StatusNotFound
		case "unauthorized to update this leave request", "only HR or manager can update leave request status":
			statusCode = fiber.StatusForbidden
		case "start date cannot be after end date", "cannot update leave request to past dates", "overlapping approved leave request exists for this date range":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Update Failed",
			Message: message,
		})
	}

	logrus.WithField("leave_request_id", id).Info("Leave request updated successfully")
	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request updated successfully",
		Data:    leaveRequest,
	})
}

func (h *LeaveRequestHandler) DeleteLeaveRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid leave request ID",
		})
	}

	// Get user info from JWT middleware
	userID := c.Locals("user_id").(uint)
	var userRole string
	if role := c.Locals("role"); role != nil {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	}

	err = h.service.DeleteLeaveRequest(uint(id), userID, userRole)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete leave request")
		statusCode := fiber.StatusInternalServerError
		message := err.Error()

		switch message {
		case "leave request not found":
			statusCode = fiber.StatusNotFound
		case "unauthorized to delete this leave request":
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Delete Failed",
			Message: message,
		})
	}

	logrus.WithField("leave_request_id", id).Info("Leave request deleted successfully")
	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request deleted successfully",
	})
}

func (h *LeaveRequestHandler) ApproveLeaveRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid leave request ID",
		})
	}

	// Get user role from JWT middleware
	var userRole string
	if role := c.Locals("role"); role != nil {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	}

	leaveRequest, err := h.service.ApproveLeaveRequest(uint(id), userRole)
	if err != nil {
		logrus.WithError(err).Error("Failed to approve leave request")
		statusCode := fiber.StatusInternalServerError
		message := err.Error()

		switch message {
		case "leave request not found":
			statusCode = fiber.StatusNotFound
		case "only HR can approve leave requests":
			statusCode = fiber.StatusForbidden
		case "overlapping approved leave request exists for this date range":
			statusCode = fiber.StatusBadRequest
		}

		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Approve Failed",
			Message: message,
		})
	}

	logrus.WithField("leave_request_id", id).Info("Leave request approved successfully")
	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request approved successfully",
		Data:    leaveRequest,
	})
}

func (h *LeaveRequestHandler) RejectLeaveRequest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dtos.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid leave request ID",
		})
	}

	// Get user role from JWT middleware
	var userRole string
	if role := c.Locals("role"); role != nil {
		if roleStr, ok := role.(string); ok {
			userRole = roleStr
		}
	}

	leaveRequest, err := h.service.RejectLeaveRequest(uint(id), userRole)
	if err != nil {
		logrus.WithError(err).Error("Failed to reject leave request")
		statusCode := fiber.StatusInternalServerError
		message := err.Error()

		switch message {
		case "leave request not found":
			statusCode = fiber.StatusNotFound
		case "only HR can reject leave requests":
			statusCode = fiber.StatusForbidden
		}

		return c.Status(statusCode).JSON(dtos.ErrorResponse{
			Error:   "Reject Failed",
			Message: message,
		})
	}

	logrus.WithField("leave_request_id", id).Info("Leave request rejected successfully")
	return c.Status(fiber.StatusOK).JSON(dtos.SuccessResponse{
		Success: true,
		Message: "Leave request rejected successfully",
		Data:    leaveRequest,
	})
}
