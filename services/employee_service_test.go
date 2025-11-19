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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestCreateEmployee(t *testing.T) {
	role := "employee"

	tests := []struct {
		name      string
		request   *dtos.CreateEmployeeRequest
		mockSetup func(*mocks.MockEmployeeRepository)
		wantError bool
		checkFunc func(*dtos.EmployeeResponse)
	}{
		{
			name: "successful creation",
			request: &dtos.CreateEmployeeRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				// Check email doesn't exist
				repo.On("FindByEmail", "john@example.com").Return(nil, gorm.ErrRecordNotFound)

				// Create employee
				repo.On("Create", mock.AnythingOfType("*models.Employee")).
					Return(nil).
					Run(func(args mock.Arguments) {
						emp := args.Get(0).(*models.Employee)
						emp.ID = 1
						emp.CreatedAt = time.Now()
						emp.UpdatedAt = time.Now()
					})
			},
			wantError: false,
			checkFunc: func(resp *dtos.EmployeeResponse) {
				assert.Equal(t, "John Doe", resp.Name)
				assert.Equal(t, "john@example.com", resp.Email)
				assert.Equal(t, uint(1), resp.ID)
			},
		},
		{
			name: "email already exists",
			request: &dtos.CreateEmployeeRequest{
				Name:     "Jane Doe",
				Email:    "existing@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				existingEmployee := &models.Employee{
					ID:    1,
					Name:  "Existing User",
					Email: "existing@example.com",
				}
				repo.On("FindByEmail", "existing@example.com").Return(existingEmployee, nil)
			},
			wantError: true,
			checkFunc: nil,
		},
		{
			name: "repository error on create",
			request: &dtos.CreateEmployeeRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
				repo.On("Create", mock.AnythingOfType("*models.Employee")).
					Return(errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewEmployeeService(mockRepo)
			result, err := service.CreateEmployee(tt.request)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(result)
				}

				// Verify password was hashed
				mockRepo.AssertCalled(t, "Create", mock.MatchedBy(func(emp *models.Employee) bool {
					err := bcrypt.CompareHashAndPassword([]byte(emp.Password), []byte(tt.request.Password))
					return err == nil
				}))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetEmployeeByID(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name      string
		id        uint
		mockSetup func(*mocks.MockEmployeeRepository)
		wantError bool
		checkFunc func(*dtos.EmployeeResponse)
	}{
		{
			name: "employee found",
			id:   1,
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{
					ID:        1,
					Name:      "John Doe",
					Email:     "john@example.com",
					Password:  "hashedpassword",
					Role:      &role,
					CreatedAt: now,
					UpdatedAt: now,
				}
				repo.On("FindByID", uint(1)).Return(employee, nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.EmployeeResponse) {
				assert.Equal(t, uint(1), resp.ID)
				assert.Equal(t, "John Doe", resp.Name)
				assert.Equal(t, "john@example.com", resp.Email)
			},
		},
		{
			name: "employee not found",
			id:   999,
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			wantError: true,
			checkFunc: nil,
		},
		{
			name: "repository error",
			id:   2,
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByID", uint(2)).Return(nil, errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewEmployeeService(mockRepo)
			result, err := service.GetEmployeeByID(tt.id)

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

func TestGetEmployees(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name      string
		request   *dtos.GetEmployeesRequest
		mockSetup func(*mocks.MockEmployeeRepository)
		wantError bool
		checkFunc func(*dtos.GetEmployeesResponse)
	}{
		{
			name: "get employees with default values",
			request: &dtos.GetEmployeesRequest{
				Page:     0,
				PageSize: 0,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employees := []models.Employee{
					{ID: 1, Name: "John", Email: "john@example.com", Password: "hash", Role: &role, CreatedAt: now, UpdatedAt: now},
					{ID: 2, Name: "Jane", Email: "jane@example.com", Password: "hash", Role: &role, CreatedAt: now, UpdatedAt: now},
				}
				repo.On("FindAll", 1, 10, "", "created_at", "desc").Return(employees, int64(2), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetEmployeesResponse) {
				assert.Equal(t, 2, len(resp.Data))
				assert.Equal(t, int64(2), resp.Pagination.TotalItems)
				assert.Equal(t, 1, resp.Pagination.CurrentPage)
				assert.Equal(t, 10, resp.Pagination.PageSize)
				assert.Equal(t, 1, resp.Pagination.TotalPages)
			},
		},
		{
			name: "get employees with custom pagination",
			request: &dtos.GetEmployeesRequest{
				Page:     2,
				PageSize: 5,
				Search:   "John",
				SortBy:   "name",
				SortDir:  "asc",
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employees := []models.Employee{
					{ID: 6, Name: "Johnny", Email: "johnny@example.com", Password: "hash", Role: &role, CreatedAt: now, UpdatedAt: now},
				}
				repo.On("FindAll", 2, 5, "John", "name", "asc").Return(employees, int64(6), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetEmployeesResponse) {
				assert.Equal(t, 1, len(resp.Data))
				assert.Equal(t, int64(6), resp.Pagination.TotalItems)
				assert.Equal(t, 2, resp.Pagination.CurrentPage)
				assert.Equal(t, 5, resp.Pagination.PageSize)
				assert.Equal(t, 2, resp.Pagination.TotalPages)
			},
		},
		{
			name: "enforce max page size",
			request: &dtos.GetEmployeesRequest{
				Page:     1,
				PageSize: 150,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employees := []models.Employee{}
				repo.On("FindAll", 1, 100, "", "created_at", "desc").Return(employees, int64(0), nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.GetEmployeesResponse) {
				assert.Equal(t, 100, resp.Pagination.PageSize)
			},
		},
		{
			name: "repository error",
			request: &dtos.GetEmployeesRequest{
				Page:     1,
				PageSize: 10,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindAll", 1, 10, "", "created_at", "desc").
					Return([]models.Employee{}, int64(0), errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			tt.mockSetup(mockRepo)

			service := NewEmployeeService(mockRepo)
			result, err := service.GetEmployees(tt.request)

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
