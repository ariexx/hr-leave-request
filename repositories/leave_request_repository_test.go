package repositories

import (
	"database/sql"
	"hr-leave-request/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupLeaveRequestMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestLeaveRequestCreate(t *testing.T) {
	now := time.Now()
	reason := "Need vacation"

	tests := []struct {
		name         string
		leaveRequest *models.LeaveRequest
		mockSetup    func(sqlmock.Sqlmock)
		wantError    bool
	}{
		{
			name: "successful creation",
			leaveRequest: &models.LeaveRequest{
				EmployeeID: 1,
				StartDate:  now.Add(24 * time.Hour),
				EndDate:    now.Add(48 * time.Hour),
				Type:       "vacation",
				Status:     "pending",
				Reason:     &reason,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `leave_requests`").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "database error",
			leaveRequest: &models.LeaveRequest{
				EmployeeID: 1,
				StartDate:  now.Add(24 * time.Hour),
				EndDate:    now.Add(48 * time.Hour),
				Type:       "sick",
				Status:     "pending",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `leave_requests`").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			err := repo.Create(tt.leaveRequest)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLeaveRequestFindByID(t *testing.T) {
	now := time.Now()
	reason := "Vacation"

	tests := []struct {
		name      string
		id        uint
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
		checkFunc func(*models.LeaveRequest)
	}{
		{
			name: "found leave request",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				// First expect leave_requests query
				rows := sqlmock.NewRows([]string{"id", "employee_id", "start_date", "end_date", "type", "status", "reason", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, 1, now, now.Add(24*time.Hour), "vacation", "pending", reason, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `leave_requests`").
					WithArgs(1, 1).
					WillReturnRows(rows)

				// Then expect preload Employee query
				employeeRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hash", "employee", now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(1).
					WillReturnRows(employeeRows)
			},
			wantError: false,
			checkFunc: func(lr *models.LeaveRequest) {
				assert.Equal(t, uint(1), lr.ID)
				assert.Equal(t, uint(1), lr.EmployeeID)
				assert.Equal(t, "vacation", lr.Type)
				assert.Equal(t, "pending", lr.Status)
			},
		},
		{
			name: "leave request not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `leave_requests`").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			result, err := repo.FindByID(tt.id)

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

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLeaveRequestFindAll(t *testing.T) {
	now := time.Now()
	status := "pending"
	leaveType := "vacation"
	employeeID := uint(1)

	tests := []struct {
		name          string
		page          int
		pageSize      int
		employeeID    *uint
		status        *string
		leaveType     *string
		startDate     *time.Time
		endDate       *time.Time
		sortBy        string
		sortDir       string
		mockSetup     func(sqlmock.Sqlmock)
		expectedCount int
		expectedTotal int64
		wantError     bool
	}{
		{
			name:       "get all leave requests",
			page:       1,
			pageSize:   10,
			employeeID: nil,
			status:     nil,
			leaveType:  nil,
			startDate:  nil,
			endDate:    nil,
			sortBy:     "created_at",
			sortDir:    "desc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(2)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests`").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "employee_id", "start_date", "end_date", "type", "status", "reason", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, 1, now, now.Add(24*time.Hour), "vacation", "pending", nil, now, now, nil).
					AddRow(2, 2, now, now.Add(24*time.Hour), "sick", "approved", nil, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `leave_requests`").
					WithArgs(10).
					WillReturnRows(rows)

				employeeRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "Employee 1", "emp1@example.com", "hash", "employee", now, now, nil).
					AddRow(2, "Employee 2", "emp2@example.com", "hash", "employee", now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees` WHERE").
					WithArgs(1, 2).
					WillReturnRows(employeeRows)
			},
			expectedCount: 2,
			expectedTotal: 2,
			wantError:     false,
		},
		{
			name:       "filter by employee and status",
			page:       1,
			pageSize:   10,
			employeeID: &employeeID,
			status:     &status,
			leaveType:  &leaveType,
			startDate:  nil,
			endDate:    nil,
			sortBy:     "created_at",
			sortDir:    "desc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests` WHERE").
					WithArgs(employeeID, status, leaveType).
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "employee_id", "start_date", "end_date", "type", "status", "reason", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, 1, now, now.Add(24*time.Hour), "vacation", "pending", nil, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `leave_requests` WHERE").
					WithArgs(employeeID, status, leaveType, 10).
					WillReturnRows(rows)

				employeeRows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hash", "employee", now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(1).
					WillReturnRows(employeeRows)
			},
			expectedCount: 1,
			expectedTotal: 1,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			results, total, err := repo.FindAll(tt.page, tt.pageSize, tt.employeeID, tt.status, tt.leaveType, tt.startDate, tt.endDate, tt.sortBy, tt.sortDir)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(results))
				assert.Equal(t, tt.expectedTotal, total)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLeaveRequestUpdate(t *testing.T) {
	now := time.Now()
	reason := "Updated reason"

	tests := []struct {
		name         string
		leaveRequest *models.LeaveRequest
		mockSetup    func(sqlmock.Sqlmock)
		wantError    bool
	}{
		{
			name: "successful update",
			leaveRequest: &models.LeaveRequest{
				ID:         1,
				EmployeeID: 1,
				StartDate:  now.Add(24 * time.Hour),
				EndDate:    now.Add(48 * time.Hour),
				Type:       "vacation",
				Status:     "approved",
				Reason:     &reason,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `leave_requests` SET").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "database error on update",
			leaveRequest: &models.LeaveRequest{
				ID:   2,
				Type: "sick",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `leave_requests` SET").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			err := repo.Update(tt.leaveRequest)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestLeaveRequestDelete(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful delete",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `leave_requests` SET `deleted_at`").
					WithArgs(sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "database error on delete",
			id:   2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `leave_requests` SET `deleted_at`").
					WithArgs(sqlmock.AnyArg(), 2).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			err := repo.Delete(tt.id)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestHasOverlappingApprovedLeave(t *testing.T) {
	now := time.Now()
	startDate := now.Add(24 * time.Hour)
	endDate := now.Add(48 * time.Hour)
	excludeID := uint(5)

	tests := []struct {
		name       string
		employeeID uint
		startDate  time.Time
		endDate    time.Time
		excludeID  *uint
		mockSetup  func(sqlmock.Sqlmock)
		wantResult bool
		wantError  bool
	}{
		{
			name:       "has overlapping leave",
			employeeID: 1,
			startDate:  startDate,
			endDate:    endDate,
			excludeID:  nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests`").
					WithArgs(1, "approved", endDate, endDate, startDate, startDate, startDate, endDate).
					WillReturnRows(countRows)
			},
			wantResult: true,
			wantError:  false,
		},
		{
			name:       "no overlapping leave",
			employeeID: 1,
			startDate:  startDate,
			endDate:    endDate,
			excludeID:  nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests`").
					WithArgs(1, "approved", endDate, endDate, startDate, startDate, startDate, endDate).
					WillReturnRows(countRows)
			},
			wantResult: false,
			wantError:  false,
		},
		{
			name:       "exclude specific request",
			employeeID: 1,
			startDate:  startDate,
			endDate:    endDate,
			excludeID:  &excludeID,
			mockSetup: func(mock sqlmock.Sqlmock) {
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(0)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests`").
					WithArgs(1, "approved", endDate, endDate, startDate, startDate, startDate, endDate, excludeID).
					WillReturnRows(countRows)
			},
			wantResult: false,
			wantError:  false,
		},
		{
			name:       "database error",
			employeeID: 1,
			startDate:  startDate,
			endDate:    endDate,
			excludeID:  nil,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `leave_requests`").
					WillReturnError(sql.ErrConnDone)
			},
			wantResult: false,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupLeaveRequestMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewLeaveRequestRepository(db)
			result, err := repo.HasOverlappingApprovedLeave(tt.employeeID, tt.startDate, tt.endDate, tt.excludeID)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
