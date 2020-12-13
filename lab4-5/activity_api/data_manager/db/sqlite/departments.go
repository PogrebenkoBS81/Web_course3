package sqlite

import (
	"activity_api/common/models"
	"database/sql"
	"errors"
	"fmt"
)

// CreateDepartment - writes given department record to SQLite db.
func (s *SQLite) CreateDepartment(depart *models.Department) (int64, error) {
	entry := s.logger.WithField("func", "CreateDepartment")

	entry.Debugf("Creating department: %+v", depart)
	result, err := s.Exec(departmentCreate, depart.DepartmentName)

	if err != nil {
		return -1, fmt.Errorf("SQLite s.Exec(), departmentCreate: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(), activityCreate: %w", err)
	}

	entry.Debugf("Created department id: %d", id)
	return id, nil
}

// GetDepartments - returns all department records from SQLite db.
func (s *SQLite) GetDepartments() ([]*models.Department, error) {
	entry := s.logger.WithField("func", "GetDepartments")

	entry.Debug("Getting departments...")
	departments := make([]*models.Department, 0)

	if err := s.Get(&departments, departmentsGet); err != nil {
		return nil, fmt.Errorf("SQLite s.Get(), departmentCreate: %w", err)
	}

	entry.Debugf("Retrieved departments num: %d", len(departments))
	return departments, nil
}

// GetDepartment - returns department record with given ID from SQLite db.
func (s *SQLite) GetDepartment(departID string) (*models.Department, error) {
	entry := s.logger.WithField("func", "GetDepartment")

	entry.Debugf("Getting department with id: %d", departID)
	department := new(models.Department)

	if err := s.Pick(department, departmentGet, departID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("SQLite s.Pick(), departmentCreate: %w", err)
	}

	entry.Debugf("Retrieved department with id %d: %+v", departID, *department)
	return department, nil
}

// DeleteDepartment - deletes department record with given ID from SQLite db.
func (s *SQLite) DeleteDepartment(departID string) (int64, error) {
	entry := s.logger.WithField("func", "DeleteDepartment")

	entry.Debugf("Deleting department with id: %s", departID)
	result, err := s.Exec(departmentDelete, departID)

	if err != nil {
		return -1, fmt.Errorf("SQLite s.Exec(), departmentCreate: %w", err)
	}

	id, err := result.RowsAffected()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(), activityCreate: %w", err)
	}

	entry.Debugf("Department with id %s deleted successfully, rows affected: %d", departID, id)
	return id, nil
}
