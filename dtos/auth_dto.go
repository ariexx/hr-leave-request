package dtos

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Email    string  `json:"email" validate:"required,email,max=100"`
	Password string  `json:"password" validate:"required,min=6,max=255"`
	Role     *string `json:"role" validate:"omitempty,oneof=employee hr manager"`
}

type AuthResponse struct {
	Token string           `json:"token"`
	User  EmployeeResponse `json:"user"`
}
