package services

import (
	"errors"
	"hr-leave-request/dtos"
	"hr-leave-request/models"
	"hr-leave-request/repositories"
	"math"
	"time"

	"gorm.io/gorm"
)

type LeaveRequestService interface {
	CreateLeaveRequest(employeeID uint, req *dtos.CreateLeaveRequestRequest) (*dtos.LeaveRequestResponse, error)
	GetLeaveRequestByID(id uint) (*dtos.LeaveRequestResponse, error)
	GetLeaveRequests(req *dtos.GetLeaveRequestsRequest) (*dtos.GetLeaveRequestsResponse, error)
	UpdateLeaveRequest(id uint, employeeID uint, userRole string, req *dtos.UpdateLeaveRequestRequest) (*dtos.LeaveRequestResponse, error)
	DeleteLeaveRequest(id uint, employeeID uint, userRole string) error
}

type leaveRequestService struct {
	repo         repositories.LeaveRequestRepository
	employeeRepo repositories.EmployeeRepository
}

func NewLeaveRequestService(repo repositories.LeaveRequestRepository, employeeRepo repositories.EmployeeRepository) LeaveRequestService {
	return &leaveRequestService{
		repo:         repo,
		employeeRepo: employeeRepo,
	}
}

func (s *leaveRequestService) CreateLeaveRequest(employeeID uint, req *dtos.CreateLeaveRequestRequest) (*dtos.LeaveRequestResponse, error) {
	// Validate employee exists
	_, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		return nil, err
	}

	// Validate date range
	if req.StartDate.After(req.EndDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	// Validate not in the past (check if start date is before current time)
	now := time.Now()
	if req.StartDate.Before(now) {
		return nil, errors.New("cannot create leave request for past dates")
	}

	// Check for overlapping approved leave requests
	hasOverlap, err := s.repo.HasOverlappingApprovedLeave(employeeID, req.StartDate, req.EndDate, nil)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, errors.New("overlapping approved leave request exists for this date range")
	}

	leaveRequest := &models.LeaveRequest{
		EmployeeID: employeeID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Type:       req.Type,
		Status:     "pending",
		Reason:     req.Reason,
	}

	if err := s.repo.Create(leaveRequest); err != nil {
		return nil, err
	}

	// Reload to get employee data
	leaveRequest, err = s.repo.FindByID(leaveRequest.ID)
	if err != nil {
		return nil, err
	}

	return s.toLeaveRequestResponse(leaveRequest), nil
}

func (s *leaveRequestService) GetLeaveRequestByID(id uint) (*dtos.LeaveRequestResponse, error) {
	leaveRequest, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("leave request not found")
		}
		return nil, err
	}

	return s.toLeaveRequestResponse(leaveRequest), nil
}

func (s *leaveRequestService) GetLeaveRequests(req *dtos.GetLeaveRequestsRequest) (*dtos.GetLeaveRequestsResponse, error) {
	// Set default values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	leaveRequests, total, err := s.repo.FindAll(
		req.Page,
		req.PageSize,
		req.EmployeeID,
		req.Status,
		req.Type,
		req.StartDate,
		req.EndDate,
		req.SortBy,
		req.SortDir,
	)
	if err != nil {
		return nil, err
	}

	leaveRequestResponses := make([]dtos.LeaveRequestResponse, len(leaveRequests))
	for i, lr := range leaveRequests {
		leaveRequestResponses[i] = *s.toLeaveRequestResponse(&lr)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dtos.GetLeaveRequestsResponse{
		Data: leaveRequestResponses,
		Pagination: dtos.PaginationMetadata{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalPages:  totalPages,
			TotalItems:  total,
		},
	}, nil
}

func (s *leaveRequestService) UpdateLeaveRequest(id uint, employeeID uint, userRole string, req *dtos.UpdateLeaveRequestRequest) (*dtos.LeaveRequestResponse, error) {
	// Get existing leave request
	leaveRequest, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("leave request not found")
		}
		return nil, err
	}

	// Authorization check: only owner or HR/manager can update
	if leaveRequest.EmployeeID != employeeID && userRole != "hr" && userRole != "manager" {
		return nil, errors.New("unauthorized to update this leave request")
	}

	// Only HR/manager can update status
	if req.Status != nil && userRole != "hr" && userRole != "manager" {
		return nil, errors.New("only HR or manager can update leave request status")
	}

	// Update fields if provided
	if req.StartDate != nil {
		leaveRequest.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		leaveRequest.EndDate = *req.EndDate
	}
	if req.Type != nil {
		leaveRequest.Type = *req.Type
	}
	if req.Status != nil {
		leaveRequest.Status = *req.Status
	}
	if req.Reason != nil {
		leaveRequest.Reason = req.Reason
	}

	// Validate date range
	if leaveRequest.StartDate.After(leaveRequest.EndDate) {
		return nil, errors.New("start date cannot be after end date")
	}

	// Validate not in the past (only if dates are being changed)
	if req.StartDate != nil || req.EndDate != nil {
		now := time.Now()
		if leaveRequest.StartDate.Before(now) {
			return nil, errors.New("cannot update leave request to past dates")
		}

		// Check for overlapping approved leave (exclude current request)
		hasOverlap, err := s.repo.HasOverlappingApprovedLeave(leaveRequest.EmployeeID, leaveRequest.StartDate, leaveRequest.EndDate, &id)
		if err != nil {
			return nil, err
		}
		if hasOverlap {
			return nil, errors.New("overlapping approved leave request exists for this date range")
		}
	}

	if err := s.repo.Update(leaveRequest); err != nil {
		return nil, err
	}

	// Reload to get updated employee data
	leaveRequest, err = s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toLeaveRequestResponse(leaveRequest), nil
}

func (s *leaveRequestService) DeleteLeaveRequest(id uint, employeeID uint, userRole string) error {
	// Get existing leave request
	leaveRequest, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("leave request not found")
		}
		return err
	}

	// Authorization check: only owner or HR/manager can delete
	if leaveRequest.EmployeeID != employeeID && userRole != "hr" && userRole != "manager" {
		return errors.New("unauthorized to delete this leave request")
	}

	return s.repo.Delete(id)
}

func (s *leaveRequestService) toLeaveRequestResponse(lr *models.LeaveRequest) *dtos.LeaveRequestResponse {
	response := &dtos.LeaveRequestResponse{
		ID:         lr.ID,
		EmployeeID: lr.EmployeeID,
		StartDate:  lr.StartDate,
		EndDate:    lr.EndDate,
		Type:       lr.Type,
		Status:     lr.Status,
		Reason:     lr.Reason,
		CreatedAt:  lr.CreatedAt,
		UpdatedAt:  lr.UpdatedAt,
	}

	if lr.Employee != nil {
		response.Employee = &dtos.EmployeeResponse{
			ID:        lr.Employee.ID,
			Name:      lr.Employee.Name,
			Email:     lr.Employee.Email,
			Role:      lr.Employee.Role,
			CreatedAt: lr.Employee.CreatedAt,
			UpdatedAt: lr.Employee.UpdatedAt,
		}
	}

	return response
}
