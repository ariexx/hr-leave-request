package services

import (
	"errors"
	"hr-leave-request/config"
	"hr-leave-request/dtos"
	"hr-leave-request/models"
	"hr-leave-request/repositories/mocks"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestConfig() *config.ApplicationConfig {
	return &config.ApplicationConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 24,
		},
	}
}

func TestLogin(t *testing.T) {
	role := "employee"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		request   *dtos.LoginRequest
		mockSetup func(*mocks.MockEmployeeRepository)
		wantError bool
		checkFunc func(*dtos.AuthResponse)
	}{
		{
			name: "successful login",
			request: &dtos.LoginRequest{
				Email:    "john@example.com",
				Password: password,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{
					ID:        1,
					Name:      "John Doe",
					Email:     "john@example.com",
					Password:  string(hashedPassword),
					Role:      &role,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repo.On("FindByEmail", "john@example.com").Return(employee, nil)
			},
			wantError: false,
			checkFunc: func(resp *dtos.AuthResponse) {
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, "John Doe", resp.User.Name)
				assert.Equal(t, "john@example.com", resp.User.Email)
				assert.Equal(t, uint(1), resp.User.ID)

				// Verify token is valid
				token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret-key"), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				claims, ok := token.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, float64(1), claims["user_id"])
				assert.Equal(t, "john@example.com", claims["email"])
			},
		},
		{
			name: "user not found",
			request: &dtos.LoginRequest{
				Email:    "notfound@example.com",
				Password: password,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByEmail", "notfound@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			wantError: true,
			checkFunc: nil,
		},
		{
			name: "invalid password",
			request: &dtos.LoginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				employee := &models.Employee{
					ID:       1,
					Name:     "John Doe",
					Email:    "john@example.com",
					Password: string(hashedPassword),
					Role:     &role,
				}
				repo.On("FindByEmail", "john@example.com").Return(employee, nil)
			},
			wantError: true,
			checkFunc: nil,
		},
		{
			name: "database error",
			request: &dtos.LoginRequest{
				Email:    "john@example.com",
				Password: password,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByEmail", "john@example.com").Return(nil, errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			cfg := setupTestConfig()
			tt.mockSetup(mockRepo)

			service := NewAuthService(mockRepo, cfg)
			result, err := service.Login(tt.request)

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

func TestRegister(t *testing.T) {
	role := "employee"

	tests := []struct {
		name      string
		request   *dtos.RegisterRequest
		mockSetup func(*mocks.MockEmployeeRepository)
		wantError bool
		checkFunc func(*dtos.AuthResponse)
	}{
		{
			name: "successful registration",
			request: &dtos.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				// Email doesn't exist
				repo.On("FindByEmail", "john@example.com").Return(nil, gorm.ErrRecordNotFound)

				// Create employee
				repo.On("Create", mock.MatchedBy(func(emp *models.Employee) bool {
					return emp.Email == "john@example.com"
				})).Return(nil).Run(func(args mock.Arguments) {
					// Simulate database setting ID and timestamps
					emp := args.Get(0).(*models.Employee)
					emp.ID = 1
					emp.CreatedAt = time.Now()
					emp.UpdatedAt = time.Now()
				})
			},
			wantError: false,
			checkFunc: func(resp *dtos.AuthResponse) {
				assert.NotEmpty(t, resp.Token)
				assert.Equal(t, "John Doe", resp.User.Name)
				assert.Equal(t, "john@example.com", resp.User.Email)

				// Verify token is valid
				token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
					return []byte("test-secret-key"), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)
			},
		},
		{
			name: "email already exists",
			request: &dtos.RegisterRequest{
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
			name: "database error on email check",
			request: &dtos.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByEmail", "test@example.com").Return(nil, errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
		{
			name: "database error on create",
			request: &dtos.RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Role:     &role,
			},
			mockSetup: func(repo *mocks.MockEmployeeRepository) {
				repo.On("FindByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
				repo.On("Create", mock.MatchedBy(func(emp *models.Employee) bool {
					return emp.Email == "test@example.com"
				})).Return(errors.New("database error"))
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			cfg := setupTestConfig()
			tt.mockSetup(mockRepo)

			service := NewAuthService(mockRepo, cfg)
			result, err := service.Register(tt.request)

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

func TestGenerateToken(t *testing.T) {
	role := "employee"
	cfg := setupTestConfig()

	tests := []struct {
		name      string
		employee  *models.Employee
		checkFunc func(string)
	}{
		{
			name: "generate valid token",
			employee: &models.Employee{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  &role,
			},
			checkFunc: func(tokenString string) {
				assert.NotEmpty(t, tokenString)

				// Parse and verify token
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					// Verify signing method
					_, ok := token.Method.(*jwt.SigningMethodHMAC)
					assert.True(t, ok)
					return []byte(cfg.JWT.Secret), nil
				})

				assert.NoError(t, err)
				assert.True(t, token.Valid)

				// Verify claims
				claims, ok := token.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, float64(1), claims["user_id"])
				assert.Equal(t, "john@example.com", claims["email"])
				assert.Equal(t, role, claims["role"])

				// Verify expiration is set
				exp, ok := claims["exp"].(float64)
				assert.True(t, ok)
				assert.Greater(t, exp, float64(time.Now().Unix()))

				// Verify issued at is set
				iat, ok := claims["iat"].(float64)
				assert.True(t, ok)
				assert.LessOrEqual(t, iat, float64(time.Now().Unix()))
			},
		},
		{
			name: "generate token with nil role",
			employee: &models.Employee{
				ID:    2,
				Name:  "Jane Doe",
				Email: "jane@example.com",
				Role:  nil,
			},
			checkFunc: func(tokenString string) {
				assert.NotEmpty(t, tokenString)

				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					return []byte(cfg.JWT.Secret), nil
				})

				assert.NoError(t, err)
				assert.True(t, token.Valid)

				claims, ok := token.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Nil(t, claims["role"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockEmployeeRepository)
			service := NewAuthService(mockRepo, cfg).(*authService)

			token, err := service.generateToken(tt.employee)

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				tt.checkFunc(token)
			}
		})
	}
}
