package sqlite

import (
	"activity_api/common/models"
	"database/sql"
	"errors"
	"fmt"
)

// CreateActivity - writes given user record to SQLite db.
func (s *SQLite) CreateUser(user *models.User) (int64, error) {
	entry := s.logger.WithField("func", "CreateUser")

	entry.Debugf("Creating user: %+v", user)
	result, err := s.Exec(userCreate, user.UserName, user.DepartmentID)

	if err != nil {
		return -1, fmt.Errorf("SQLite Exec(), userCreate: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(), activityCreate: %w", err)
	}

	entry.Debugf("Created user id: %d", id)
	return id, nil
}

// GetUsers - returns all users records from SQLite db.
func (s *SQLite) GetUsers(depID string) ([]*models.User, error) {
	entry := s.logger.WithField("func", "GetUsers")

	entry.Debugf("Getting users, department id: %s", depID)
	users := make([]*models.User, 0)

	if depID == "" {
		if err := s.Get(&users, userDepartmentGet, depID); err != nil {
			return nil, fmt.Errorf("s.Get() userDepartmentGet : %w", err)
		}
	} else {
		if err := s.Get(&users, usersGet); err != nil {
			return nil, fmt.Errorf("s.Get() userGet : %w", err)
		}
	}

	entry.Debugf("Retrieved users: %+v", users)
	return users, nil
}

// GetUser - returns user record with given ID from SQLite db.
func (s *SQLite) GetUser(userID string) (*models.User, error) {
	entry := s.logger.WithField("func", "GetUser")

	entry.Debugf("Getting user with id: %s", userID)
	user := new(models.User)

	if err := s.Pick(user, userGet, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("conn.QueryRow() userGet : %w", err)
	}

	entry.Debugf("Retrieved users with id %s: %+v", userID, user)
	return user, nil
}

// DeleteUser - deletes user record with given ID from SQLite db.
func (s *SQLite) DeleteUser(userID string) (int64, error) {
	entry := s.logger.WithField("func", "DeleteUser")

	entry.Debugf("Deleting user with id: %s", userID)
	result, err := s.Exec(userDelete, userID)

	if err != nil {
		return -1, fmt.Errorf("SQLite Exec(), userDelete: %w", err)
	}

	id, err := result.RowsAffected()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(), activityCreate: %w", err)
	}

	entry.Debugf("User with id %s deleted successfully, rows affected: %d", userID, id)
	return id, nil
}
