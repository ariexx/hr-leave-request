package repositories

import (
	"hr-leave-request/models"
	"time"

	"gorm.io/gorm"
)

type LeaveRequestRepository interface {
	Create(leaveRequest *models.LeaveRequest) error
	FindByID(id uint) (*models.LeaveRequest, error)
	FindAll(page, pageSize int, employeeID *uint, status, leaveType *string, startDate, endDate *time.Time, sortBy, sortDir string) ([]models.LeaveRequest, int64, error)
	Update(leaveRequest *models.LeaveRequest) error
	Delete(id uint) error
	HasOverlappingApprovedLeave(employeeID uint, startDate, endDate time.Time, excludeID *uint) (bool, error)
}

type leaveRequestRepository struct {
	db *gorm.DB
}

func NewLeaveRequestRepository(db *gorm.DB) LeaveRequestRepository {
	return &leaveRequestRepository{db: db}
}

func (r *leaveRequestRepository) Create(leaveRequest *models.LeaveRequest) error {
	return r.db.Create(leaveRequest).Error
}

func (r *leaveRequestRepository) FindByID(id uint) (*models.LeaveRequest, error) {
	var leaveRequest models.LeaveRequest
	err := r.db.Preload("Employee").First(&leaveRequest, id).Error
	if err != nil {
		return nil, err
	}
	return &leaveRequest, nil
}

func (r *leaveRequestRepository) FindAll(page, pageSize int, employeeID *uint, status, leaveType *string, startDate, endDate *time.Time, sortBy, sortDir string) ([]models.LeaveRequest, int64, error) {
	var leaveRequests []models.LeaveRequest
	var total int64

	query := r.db.Model(&models.LeaveRequest{})

	// Apply filters
	if employeeID != nil {
		query = query.Where("employee_id = ?", *employeeID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if leaveType != nil {
		query = query.Where("type = ?", *leaveType)
	}
	if startDate != nil {
		query = query.Where("start_date >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("end_date <= ?", *endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	offset := (page - 1) * pageSize
	query = query.Order(sortBy + " " + sortDir).Limit(pageSize).Offset(offset)

	// Preload employee data
	query = query.Preload("Employee")

	if err := query.Find(&leaveRequests).Error; err != nil {
		return nil, 0, err
	}

	return leaveRequests, total, nil
}

func (r *leaveRequestRepository) Update(leaveRequest *models.LeaveRequest) error {
	return r.db.Save(leaveRequest).Error
}

func (r *leaveRequestRepository) Delete(id uint) error {
	return r.db.Delete(&models.LeaveRequest{}, id).Error
}

// HasOverlappingApprovedLeave checks if there are any approved leave requests
// that overlap with the given date range for the specified employee
// excludeID is used to exclude a specific leave request (useful for updates)
func (r *leaveRequestRepository) HasOverlappingApprovedLeave(employeeID uint, startDate, endDate time.Time, excludeID *uint) (bool, error) {
	var count int64

	query := r.db.Model(&models.LeaveRequest{}).
		Where("employee_id = ?", employeeID).
		Where("status = ?", "approved").
		Where(
			"(start_date <= ? AND end_date >= ?) OR "+
				"(start_date <= ? AND end_date >= ?) OR "+
				"(start_date >= ? AND end_date <= ?)",
			endDate, endDate, // New request ends during existing leave
			startDate, startDate, // New request starts during existing leave
			startDate, endDate, // Existing leave is completely within new request
		)

	// Exclude a specific leave request if provided (for update operations)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
