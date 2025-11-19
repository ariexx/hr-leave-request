package mocks

import (
	"hr-leave-request/models"

	"github.com/stretchr/testify/mock"
)

type MockEmployeeRepository struct {
	mock.Mock
}

func (m *MockEmployeeRepository) Create(employee *models.Employee) error {
	args := m.Called(employee)
	return args.Error(0)
}

func (m *MockEmployeeRepository) FindByID(id uint) (*models.Employee, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindByEmail(email string) (*models.Employee, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Employee), args.Error(1)
}

func (m *MockEmployeeRepository) FindAll(page, pageSize int, search, sortBy, sortDir string) ([]models.Employee, int64, error) {
	args := m.Called(page, pageSize, search, sortBy, sortDir)
	return args.Get(0).([]models.Employee), args.Get(1).(int64), args.Error(2)
}

func (m *MockEmployeeRepository) Update(employee *models.Employee) error {
	args := m.Called(employee)
	return args.Error(0)
}

func (m *MockEmployeeRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}
