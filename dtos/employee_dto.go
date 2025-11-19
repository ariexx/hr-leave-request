package dtos

import "time"

type CreateEmployeeRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=100"`
	Email    string  `json:"email" validate:"required,email,max=100"`
	Password string  `json:"password" validate:"required,min=6,max=255"`
	Role     *string `json:"role" validate:"omitempty,oneof=employee hr manager"`
}

type EmployeeResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      *string   `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetEmployeesRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	Search   string `query:"search"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=name email created_at"`
	SortDir  string `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type PaginationMetadata struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
}

type GetEmployeesResponse struct {
	Data       []EmployeeResponse `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type SuccessResponsePaginated struct {
	Success    bool               `json:"success"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data,omitempty"`
	Pagination PaginationMetadata `json:"pagination"`
}
