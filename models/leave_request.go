package models

import (
	"time"

	"gorm.io/gorm"
)

type LeaveRequest struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	EmployeeID uint           `gorm:"not null;index" json:"employee_id"`
	Employee   *Employee      `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
	StartDate  time.Time      `gorm:"type:datetime;not null" json:"start_date"`
	EndDate    time.Time      `gorm:"type:datetime;not null" json:"end_date"`
	Type       string         `gorm:"type:enum('sick','vacation','personal','other');not null" json:"type"`
	Status     string         `gorm:"type:enum('pending','approved','rejected');not null;default:'pending'" json:"status"`
	Reason     *string        `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (LeaveRequest) TableName() string {
	return "leave_requests"
}
