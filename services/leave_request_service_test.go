package services

import (
	"errors"
	"hr-leave-request/dtos"
	"hr-leave-request/models"
	"hr-leave-request/repositories/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCreateLeaveRequest(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	futureEnd := now.Add(72 * time.Hour)
	reason := "Need vacation"

	tests := []struct {
		name       string
		employeeID uint
		request    *dtos.CreateLeaveRequestRequest
		mockSetup  func(*mocks.MockLeaveRequestRepository, *mocks.MockEmployeeRepository)
		wantError  bool
		errorMsg   string
		checkFunc  func(*dtos.LeaveRequestResponse)
	}{
		{
			name:       "successful creation",
			employeeID: 1,
			request: &dtos.CreateLeaveRequestRequest{
				StartDate: future,
				EndDate:   futureEnd,
				Type:      "vacation",
				Reason:    &reason,
			},
			mockSetup: func(leaveRepo *mocks.MockLeaveRequestRepository, empRepo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				empRepo.On("FindByID", uint(1)).Return(employee, nil)
				leaveRepo.On("HasOverlappingApprovedLeave", uint(1), future, futureEnd, (*uint)(nil)).Return(false, nil)
				leaveRepo.On("Create", mock.AnythingOfType("*models.LeaveRequest")).
					Return(nil).
					Run(func(args mock.Arguments) {
						lr := args.Get(0).(*models.LeaveRequest)
						lr.ID = 1
						lr.CreatedAt = now
						lr.UpdatedAt = now
					})

				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Employee:   employee,
					StartDate:  future,
					EndDate:    futureEnd,
					Type:       "vacation",
					Status:     "pending",
					Reason:     &reason,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				leaveRepo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.LeaveRequestResponse) {
				assert.Equal(t, uint(1), resp.ID)
				assert.Equal(t, uint(1), resp.EmployeeID)
				assert.Equal(t, "vacation", resp.Type)
				assert.Equal(t, "pending", resp.Status)
			},
		},
		{
			name:       "employee not found",
			employeeID: 999,
			request: &dtos.CreateLeaveRequestRequest{
				StartDate: future,
				EndDate:   futureEnd,
				Type:      "vacation",
			},
			mockSetup: func(leaveRepo *mocks.MockLeaveRequestRepository, empRepo *mocks.MockEmployeeRepository) {
				empRepo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			wantError: true,
			errorMsg:  "employee not found",
		},
		{
			name:       "start date after end date",
			employeeID: 1,
			request: &dtos.CreateLeaveRequestRequest{
				StartDate: futureEnd,
				EndDate:   future,
				Type:      "vacation",
			},
			mockSetup: func(leaveRepo *mocks.MockLeaveRequestRepository, empRepo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{ID: 1}
				empRepo.On("FindByID", uint(1)).Return(employee, nil)
			},
			wantError: true,
			errorMsg:  "start date cannot be after end date",
		},
		{
			name:       "start date in the past",
			employeeID: 1,
			request: &dtos.CreateLeaveRequestRequest{
				StartDate: now.Add(-24 * time.Hour),
				EndDate:   now.Add(24 * time.Hour),
				Type:      "vacation",
			},
			mockSetup: func(leaveRepo *mocks.MockLeaveRequestRepository, empRepo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{ID: 1}
				empRepo.On("FindByID", uint(1)).Return(employee, nil)
			},
			wantError: true,
			errorMsg:  "cannot create leave request for past dates",
		},
		{
			name:       "overlapping approved leave exists",
			employeeID: 1,
			request: &dtos.CreateLeaveRequestRequest{
				StartDate: future,
				EndDate:   futureEnd,
				Type:      "vacation",
			},
			mockSetup: func(leaveRepo *mocks.MockLeaveRequestRepository, empRepo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{ID: 1}
				empRepo.On("FindByID", uint(1)).Return(employee, nil)
				leaveRepo.On("HasOverlappingApprovedLeave", uint(1), future, futureEnd, (*uint)(nil)).Return(true, nil)
			},
			wantError: true,
			errorMsg:  "overlapping approved leave request exists for this date range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLeaveRepo := new(mocks.MockLeaveRequestRepository)
			mockEmpRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockLeaveRepo, mockEmpRepo)

			service := NewLeaveRequestService(mockLeaveRepo, mockEmpRepo)
			result, err := service.CreateLeaveRequest(tt.employeeID, tt.request)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(result)
				}
			}

			mockLeaveRepo.AssertExpectations(t)
			mockEmpRepo.AssertExpectations(t)
		})
	}
}

func TestGetLeaveRequestByID(t *testing.T) {
	now := time.Now()
	reason := "Vacation"

	tests := []struct {
		name      string
		id        uint
		mockSetup func(*mocks.MockLeaveRequestRepository)
		wantError bool
		checkFunc func(*dtos.LeaveRequestResponse)
	}{
		{
			name: "leave request found",
			id:   1,
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				employee := &models.Employee{
					ID:    1,
					Name:  "John Doe",
					Email: "john@example.com",
				}
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Employee:   employee,
					StartDate:  now,
					EndDate:    now.Add(24 * time.Hour),
					Type:       "vacation",
					Status:     "pending",
					Reason:     &reason,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.LeaveRequestResponse) {
				assert.Equal(t, uint(1), resp.ID)
				assert.Equal(t, "vacation", resp.Type)
				assert.NotNil(t, resp.Employee)
				assert.Equal(t, "John Doe", resp.Employee.Name)
			},
		},
		{
			name: "leave request not found",
			id:   999,
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				repo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockLeaveRequestRepository)
			mockEmpRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewLeaveRequestService(mockRepo, mockEmpRepo)
			result, err := service.GetLeaveRequestByID(tt.id)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetLeaveRequests(t *testing.T) {
	now := time.Now()
	status := "pending"
	empID := uint(1)

	tests := []struct {
		name      string
		request   *dtos.GetLeaveRequestsRequest
		mockSetup func(*mocks.MockLeaveRequestRepository)
		wantError bool
		checkFunc func(*dtos.GetLeaveRequestsResponse)
	}{
		{
			name: "get leave requests with default values",
			request: &dtos.GetLeaveRequestsRequest{
				Page:     0,
				PageSize: 0,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequests := []models.LeaveRequest{
					{ID: 1, EmployeeID: 1, Type: "vacation", Status: "pending", CreatedAt: now},
					{ID: 2, EmployeeID: 2, Type: "sick", Status: "approved", CreatedAt: now},
				}
				repo.On("FindAll", 1, 10, (*uint)(nil), (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), "created_at", "desc").
					Return(leaveRequests, int64(2), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetLeaveRequestsResponse) {
				assert.Equal(t, 2, len(resp.Data))
				assert.Equal(t, int64(2), resp.Pagination.TotalItems)
				assert.Equal(t, 1, resp.Pagination.CurrentPage)
				assert.Equal(t, 10, resp.Pagination.PageSize)
			},
		},
		{
			name: "filter by employee and status",
			request: &dtos.GetLeaveRequestsRequest{
				Page:       1,
				PageSize:   10,
				EmployeeID: &empID,
				Status:     &status,
				SortBy:     "start_date",
				SortDir:    "asc",
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequests := []models.LeaveRequest{
					{ID: 1, EmployeeID: 1, Type: "vacation", Status: "pending", CreatedAt: now},
				}
				repo.On("FindAll", 1, 10, &empID, &status, (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), "start_date", "asc").
					Return(leaveRequests, int64(1), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetLeaveRequestsResponse) {
				assert.Equal(t, 1, len(resp.Data))
				assert.Equal(t, int64(1), resp.Pagination.TotalItems)
			},
		},
		{
			name: "enforce max page size",
			request: &dtos.GetLeaveRequestsRequest{
				Page:     1,
				PageSize: 150,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				repo.On("FindAll", 1, 100, (*uint)(nil), (*string)(nil), (*string)(nil), (*time.Time)(nil), (*time.Time)(nil), "created_at", "desc").
					Return([]models.LeaveRequest{}, int64(0), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetLeaveRequestsResponse) {
				assert.Equal(t, 100, resp.Pagination.PageSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockLeaveRequestRepository)
			mockEmpRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewLeaveRequestService(mockRepo, mockEmpRepo)
			result, err := service.GetLeaveRequests(tt.request)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateLeaveRequest(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	futureEnd := now.Add(72 * time.Hour)
	newStatus := "approved"
	newType := "sick"

	tests := []struct {
		name       string
		id         uint
		employeeID uint
		userRole   string
		request    *dtos.UpdateLeaveRequestRequest
		mockSetup  func(*mocks.MockLeaveRequestRepository)
		wantError  bool
		errorMsg   string
	}{
		{
			name:       "successful update by owner",
			id:         1,
			employeeID: 1,
			userRole:   "employee",
			request: &dtos.UpdateLeaveRequestRequest{
				Type: &newType,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					StartDate:  future,
					EndDate:    futureEnd,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil).Once()
				repo.On("Update", mock.AnythingOfType("*models.LeaveRequest")).Return(nil)
				updatedLeaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					StartDate:  future,
					EndDate:    futureEnd,
					Type:       "sick",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(updatedLeaveRequest, nil).Once()
			},
			wantError: false,
		},
		{
			name:       "hr updates status",
			id:         1,
			employeeID: 2,
			userRole:   "hr",
			request: &dtos.UpdateLeaveRequestRequest{
				Status: &newStatus,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					StartDate:  future,
					EndDate:    futureEnd,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil).Once()
				repo.On("Update", mock.AnythingOfType("*models.LeaveRequest")).Return(nil)
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil).Once()
			},
			wantError: false,
		},
		{
			name:       "unauthorized update",
			id:         1,
			employeeID: 2,
			userRole:   "employee",
			request: &dtos.UpdateLeaveRequestRequest{
				Type: &newType,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: true,
			errorMsg:  "unauthorized to update this leave request",
		},
		{
			name:       "employee tries to update status",
			id:         1,
			employeeID: 1,
			userRole:   "employee",
			request: &dtos.UpdateLeaveRequestRequest{
				Status: &newStatus,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: true,
			errorMsg:  "only HR or manager can update leave request status",
		},
		{
			name:       "update to past dates",
			id:         1,
			employeeID: 1,
			userRole:   "employee",
			request: &dtos.UpdateLeaveRequestRequest{
				StartDate: &now,
			},
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					StartDate:  future,
					EndDate:    futureEnd,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: true,
			errorMsg:  "cannot update leave request to past dates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockLeaveRequestRepository)
			mockEmpRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewLeaveRequestService(mockRepo, mockEmpRepo)
			result, err := service.UpdateLeaveRequest(tt.id, tt.employeeID, tt.userRole, tt.request)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteLeaveRequest(t *testing.T) {
	tests := []struct {
		name       string
		id         uint
		employeeID uint
		userRole   string
		mockSetup  func(*mocks.MockLeaveRequestRepository)
		wantError  bool
		errorMsg   string
	}{
		{
			name:       "successful delete by owner",
			id:         1,
			employeeID: 1,
			userRole:   "employee",
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
				repo.On("Delete", uint(1)).Return(nil)
			},
			wantError: false,
		},
		{
			name:       "successful delete by hr",
			id:         1,
			employeeID: 2,
			userRole:   "hr",
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
				repo.On("Delete", uint(1)).Return(nil)
			},
			wantError: false,
		},
		{
			name:       "unauthorized delete",
			id:         1,
			employeeID: 2,
			userRole:   "employee",
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
			},
			wantError: true,
			errorMsg:  "unauthorized to delete this leave request",
		},
		{
			name:       "leave request not found",
			id:         999,
			employeeID: 1,
			userRole:   "employee",
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				repo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			wantError: true,
			errorMsg:  "leave request not found",
		},
		{
			name:       "database error on delete",
			id:         1,
			employeeID: 1,
			userRole:   "employee",
			mockSetup: func(repo *mocks.MockLeaveRequestRepository) {
				leaveRequest := &models.LeaveRequest{
					ID:         1,
					EmployeeID: 1,
					Type:       "vacation",
					Status:     "pending",
				}
				repo.On("FindByID", uint(1)).Return(leaveRequest, nil)
				repo.On("Delete", uint(1)).Return(errors.New("database error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockLeaveRequestRepository)
			mockEmpRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewLeaveRequestService(mockRepo, mockEmpRepo)
			err := service.DeleteLeaveRequest(tt.id, tt.employeeID, tt.userRole)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
