package dtos

import "time"

type CreateLeaveRequestRequest struct {
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	Type      string    `json:"type" validate:"required,oneof=sick vacation personal other"`
	Reason    *string   `json:"reason" validate:"omitempty"`
}

type UpdateLeaveRequestRequest struct {
	StartDate *time.Time `json:"start_date" validate:"omitempty"`
	EndDate   *time.Time `json:"end_date" validate:"omitempty,gtfield=StartDate"`
	Type      *string    `json:"type" validate:"omitempty,oneof=sick vacation personal other"`
	Status    *string    `json:"status" validate:"omitempty,oneof=pending approved rejected"`
	Reason    *string    `json:"reason" validate:"omitempty"`
}

type LeaveRequestResponse struct {
	ID         uint              `json:"id"`
	EmployeeID uint              `json:"employee_id"`
	Employee   *EmployeeResponse `json:"employee,omitempty"`
	StartDate  time.Time         `json:"start_date"`
	EndDate    time.Time         `json:"end_date"`
	Type       string            `json:"type"`
	Status     string            `json:"status"`
	Reason     *string           `json:"reason,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type GetLeaveRequestsRequest struct {
	Page       int        `query:"page" validate:"min=1"`
	PageSize   int        `query:"page_size" validate:"min=1,max=100"`
	EmployeeID *uint      `query:"employee_id"`
	Status     *string    `query:"status" validate:"omitempty,oneof=pending approved rejected"`
	Type       *string    `query:"type" validate:"omitempty,oneof=sick vacation personal other"`
	StartDate  *time.Time `query:"start_date"`
	EndDate    *time.Time `query:"end_date"`
	SortBy     string     `query:"sort_by" validate:"omitempty,oneof=start_date end_date created_at"`
	SortDir    string     `query:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

type GetLeaveRequestsResponse struct {
	Data       []LeaveRequestResponse `json:"data"`
	Pagination PaginationMetadata     `json:"pagination"`
}
