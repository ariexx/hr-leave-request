package mocks

import (
	"hr-leave-request/models"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockLeaveRequestRepository struct {
	mock.Mock
}

func (m *MockLeaveRequestRepository) Create(leaveRequest *models.LeaveRequest) error {
	args := m.Called(leaveRequest)
	return args.Error(0)
}

func (m *MockLeaveRequestRepository) FindByID(id uint) (*models.LeaveRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LeaveRequest), args.Error(1)
}

func (m *MockLeaveRequestRepository) FindAll(page, pageSize int, employeeID *uint, status, leaveType *string, startDate, endDate *time.Time, sortBy, sortDir string) ([]models.LeaveRequest, int64, error) {
	args := m.Called(page, pageSize, employeeID, status, leaveType, startDate, endDate, sortBy, sortDir)
	return args.Get(0).([]models.LeaveRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockLeaveRequestRepository) Update(leaveRequest *models.LeaveRequest) error {
	args := m.Called(leaveRequest)
	return args.Error(0)
}

func (m *MockLeaveRequestRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockLeaveRequestRepository) HasOverlappingApprovedLeave(employeeID uint, startDate, endDate time.Time, excludeID *uint) (bool, error) {
	args := m.Called(employeeID, startDate, endDate, excludeID)
	return args.Bool(0), args.Error(1)
}
