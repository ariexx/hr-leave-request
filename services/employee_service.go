package services

import (
	"errors"
	"hr-leave-request/dtos"
	"hr-leave-request/models"
	"hr-leave-request/repositories"
	"math"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type EmployeeService interface {
	CreateEmployee(req *dtos.CreateEmployeeRequest) (*dtos.EmployeeResponse, error)
	GetEmployeeByID(id uint) (*dtos.EmployeeResponse, error)
	GetEmployees(req *dtos.GetEmployeesRequest) (*dtos.GetEmployeesResponse, error)
}

type employeeService struct {
	repo repositories.EmployeeRepository
}

func NewEmployeeService(repo repositories.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) CreateEmployee(req *dtos.CreateEmployeeRequest) (*dtos.EmployeeResponse, error) {
	// Check if email already exists
	existingEmployee, err := s.repo.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingEmployee != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	employee := &models.Employee{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(employee); err != nil {
		return nil, err
	}

	return s.toEmployeeResponse(employee), nil
}

func (s *employeeService) GetEmployeeByID(id uint) (*dtos.EmployeeResponse, error) {
	employee, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("employee not found")
		}
		return nil, err
	}

	return s.toEmployeeResponse(employee), nil
}

func (s *employeeService) GetEmployees(req *dtos.GetEmployeesRequest) (*dtos.GetEmployeesResponse, error) {
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

	employees, total, err := s.repo.FindAll(req.Page, req.PageSize, req.Search, req.SortBy, req.SortDir)
	if err != nil {
		return nil, err
	}

	employeeResponses := make([]dtos.EmployeeResponse, len(employees))
	for i, emp := range employees {
		employeeResponses[i] = *s.toEmployeeResponse(&emp)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dtos.GetEmployeesResponse{
		Data: employeeResponses,
		Pagination: dtos.PaginationMetadata{
			CurrentPage: req.Page,
			PageSize:    req.PageSize,
			TotalPages:  totalPages,
			TotalItems:  total,
		},
	}, nil
}

func (s *employeeService) toEmployeeResponse(employee *models.Employee) *dtos.EmployeeResponse {
	return &dtos.EmployeeResponse{
		ID:        employee.ID,
		Name:      employee.Name,
		Email:     employee.Email,
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt,
	}
}
