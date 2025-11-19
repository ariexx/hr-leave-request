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

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
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

func TestCreate(t *testing.T) {
	role := "employee"

	tests := []struct {
		name      string
		employee  *models.Employee
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful creation",
			employee: &models.Employee{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "hashedpassword",
				Role:     &role,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `employees`").
					WithArgs("John Doe", "john@example.com", "hashedpassword", &role, sqlmock.AnyArg(), sqlmock.AnyArg(), nil).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "database error",
			employee: &models.Employee{
				Name:     "Jane Doe",
				Email:    "jane@example.com",
				Password: "hashedpassword",
				Role:     &role,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO `employees`").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
			err := repo.Create(tt.employee)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFindByID(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name      string
		id        uint
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
		checkFunc func(*models.Employee)
	}{
		{
			name: "found employee",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hashedpassword", role, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			wantError: false,
			checkFunc: func(emp *models.Employee) {
				assert.Equal(t, uint(1), emp.ID)
				assert.Equal(t, "John Doe", emp.Name)
				assert.Equal(t, "john@example.com", emp.Email)
			},
		},
		{
			name: "employee not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(999, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
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

func TestFindByEmail(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name      string
		email     string
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
		checkFunc func(*models.Employee)
	}{
		{
			name:  "found employee by email",
			email: "john@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hashedpassword", role, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees` WHERE email = \\?").
					WithArgs("john@example.com", 1).
					WillReturnRows(rows)
			},
			wantError: false,
			checkFunc: func(emp *models.Employee) {
				assert.Equal(t, "john@example.com", emp.Email)
				assert.Equal(t, "John Doe", emp.Name)
			},
		},
		{
			name:  "employee not found",
			email: "notfound@example.com",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM `employees` WHERE email = \\?").
					WithArgs("notfound@example.com", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
			result, err := repo.FindByEmail(tt.email)

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

func TestFindAll(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name          string
		page          int
		pageSize      int
		search        string
		sortBy        string
		sortDir       string
		mockSetup     func(sqlmock.Sqlmock)
		expectedCount int
		expectedTotal int64
		wantError     bool
	}{
		{
			name:     "get all employees",
			page:     1,
			pageSize: 10,
			search:   "",
			sortBy:   "created_at",
			sortDir:  "desc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(3)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `employees`").
					WillReturnRows(countRows)

				// Select query
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hash1", role, now, now, nil).
					AddRow(2, "Jane Doe", "jane@example.com", "hash2", role, now, now, nil).
					AddRow(3, "Bob Smith", "bob@example.com", "hash3", role, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(10).
					WillReturnRows(rows)
			},
			expectedCount: 3,
			expectedTotal: 3,
			wantError:     false,
		},
		{
			name:     "search by name",
			page:     1,
			pageSize: 10,
			search:   "John",
			sortBy:   "created_at",
			sortDir:  "desc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query with search
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(1)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `employees` WHERE").
					WithArgs("%John%", "%John%").
					WillReturnRows(countRows)

				// Select query with search
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(1, "John Doe", "john@example.com", "hash1", role, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees` WHERE").
					WithArgs("%John%", "%John%", 10).
					WillReturnRows(rows)
			},
			expectedCount: 1,
			expectedTotal: 1,
			wantError:     false,
		},
		{
			name:     "pagination page 2",
			page:     2,
			pageSize: 2,
			search:   "",
			sortBy:   "name",
			sortDir:  "asc",
			mockSetup: func(mock sqlmock.Sqlmock) {
				// Count query
				countRows := sqlmock.NewRows([]string{"count(*)"}).AddRow(5)
				mock.ExpectQuery("SELECT count\\(\\*\\) FROM `employees`").
					WillReturnRows(countRows)

				// Select query with offset
				rows := sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "created_at", "updated_at", "deleted_at"}).
					AddRow(3, "Charlie", "charlie@example.com", "hash3", role, now, now, nil).
					AddRow(4, "David", "david@example.com", "hash4", role, now, now, nil)
				mock.ExpectQuery("SELECT \\* FROM `employees`").
					WithArgs(2, 2).
					WillReturnRows(rows)
			},
			expectedCount: 2,
			expectedTotal: 5,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
			results, total, err := repo.FindAll(tt.page, tt.pageSize, tt.search, tt.sortBy, tt.sortDir)

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

func TestUpdate(t *testing.T) {
	role := "employee"
	now := time.Now()

	tests := []struct {
		name      string
		employee  *models.Employee
		mockSetup func(sqlmock.Sqlmock)
		wantError bool
	}{
		{
			name: "successful update",
			employee: &models.Employee{
				ID:        1,
				Name:      "Updated Name",
				Email:     "updated@example.com",
				Password:  "hashedpassword",
				Role:      &role,
				CreatedAt: now,
				UpdatedAt: now,
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `employees` SET").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantError: false,
		},
		{
			name: "database error on update",
			employee: &models.Employee{
				ID:   2,
				Name: "Failed Update",
			},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE `employees` SET").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
			err := repo.Update(tt.employee)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDelete(t *testing.T) {
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
				mock.ExpectExec("UPDATE `employees` SET `deleted_at`").
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
				mock.ExpectExec("UPDATE `employees` SET `deleted_at`").
					WithArgs(sqlmock.AnyArg(), 2).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, cleanup := setupMockDB(t)
			defer cleanup()

			tt.mockSetup(mock)

			repo := NewEmployeeRepository(db)
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
