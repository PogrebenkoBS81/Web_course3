package sqlite

import (
	"activity_api/common/models"
	"database/sql"
	"errors"
	"fmt"
)

// CreateAdmin - creates service admin with given data.
func (s *SQLite) CreateAdmin(admin *models.Admin) (int64, error) {
	entry := s.logger.WithField("func", "CreateAdmin")

	entry.Debugf("Creating admin with name: %s", admin.Username)
	result, err := s.Exec(adminCreate, admin.Username, admin.Hash)

	if err != nil {
		return -1, fmt.Errorf("SQLite Exec(): %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(): %w", err)
	}

	entry.Debugf("Created admin id: %d", id)
	return id, nil
}

// GetAdmin - returns admin with given name.
func (s *SQLite) GetAdmin(name string) (*models.Admin, error) {
	entry := s.logger.WithField("func", "GetAdmin")

	entry.Debugf("Getting admin with name %s", name)
	admin := new(models.Admin)

	if err := s.Pick(admin, adminFind, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("SQLite Pick() adminFind : %w", err)
	}

	entry.Debugf("Admin %s retrieved successfully", name)
	return admin, nil
}

// DeleteAdmin - deletes admin with given name.
func (s *SQLite) DeleteAdmin(name string) (int64, error) {
	entry := s.logger.WithField("func", "DeleteAdmin")

	entry.Debugf("Deleting admin with name: %s", name)
	result, err := s.Exec(adminDelete, name)

	if err != nil {
		return -1, fmt.Errorf("SQLite Exec(): %w", err)
	}

	id, err := result.RowsAffected()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(): %w", err)
	}

	entry.Debugf("Admin with name %s deleted successfully, rows affected: %d", name, id)
	return id, nil
}
