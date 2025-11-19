package services

import (
	"errors"
	"hr-leave-request/config"
	"hr-leave-request/dtos"
	"hr-leave-request/models"
	"hr-leave-request/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(req *dtos.LoginRequest) (*dtos.AuthResponse, error)
	Register(req *dtos.RegisterRequest) (*dtos.AuthResponse, error)
}

type authService struct {
	repo repositories.EmployeeRepository
	cfg  *config.ApplicationConfig
}

func NewAuthService(repo repositories.EmployeeRepository, cfg *config.ApplicationConfig) AuthService {
	return &authService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *authService) Login(req *dtos.LoginRequest) (*dtos.AuthResponse, error) {
	// Find user by email
	employee, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.generateToken(employee)
	if err != nil {
		return nil, err
	}

	return &dtos.AuthResponse{
		Token: token,
		User:  *s.toEmployeeResponse(employee),
	}, nil
}

func (s *authService) Register(req *dtos.RegisterRequest) (*dtos.AuthResponse, error) {
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
		Role:     req.Role,
	}

	if err := s.repo.Create(employee); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateToken(employee)
	if err != nil {
		return nil, err
	}

	return &dtos.AuthResponse{
		Token: token,
		User:  *s.toEmployeeResponse(employee),
	}, nil
}

func (s *authService) generateToken(employee *models.Employee) (string, error) {
	claims := jwt.MapClaims{
		"user_id": employee.ID,
		"email":   employee.Email,
		"role":    employee.Role,
		"exp":     time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.Expiration)).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

func (s *authService) toEmployeeResponse(employee *models.Employee) *dtos.EmployeeResponse {
	return &dtos.EmployeeResponse{
		ID:        employee.ID,
		Name:      employee.Name,
		Email:     employee.Email,
		Role:      employee.Role,
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt,
	}
}
