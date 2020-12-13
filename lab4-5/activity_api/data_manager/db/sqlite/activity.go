package sqlite

import (
	"activity_api/common/models"
	"database/sql"
	"errors"
	"fmt"
)

// CreateActivity - writes given activity record to SQLite db.
func (s *SQLite) CreateActivity(activity *models.Activity) (int64, error) {
	entry := s.logger.WithField("func", "CreateDB")
	entry.Debugf("Creating activity: %+v", activity)

	result, err := s.Exec(
		activityCreate,
		activity.UserID,
		activity.ActiveTime,
		activity.TotalTime,
		activity.Date,
	)

	if err != nil {
		return -1, fmt.Errorf("SQLite s.Exec(): %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(): %w", err)
	}

	entry.Debugf("Created activity id: %d", id)
	return id, nil
}

// GetActivity - returns all activity records from SQLite db.
func (s *SQLite) GetActivities() ([]*models.Activity, error) {
	entry := s.logger.WithField("func", "GetActivities")

	entry.Debug("Getting activities...")
	activities := make([]*models.Activity, 0)

	if err := s.Get(&activities, activitiesGet); err != nil {
		return nil, fmt.Errorf("SQLite s.Get(): %w", err)
	}

	entry.Debugf("Retrieved activities: %d", len(activities))
	return activities, nil
}

// GetActivity - returns activity record with given ID from SQLite db.
func (s *SQLite) GetActivity(activityID string) (*models.Activity, error) {
	entry := s.logger.WithField("func", "GetActivity")

	entry.Debugf("Getting activity with id: %s", activityID)
	activity := new(models.Activity)

	if err := s.Pick(activity, activityGet, activityID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("SQLite s.Pick(): %w", err)
	}

	s.logger.Debugf("Retrieved activity with id %s: %+v", activityID, *activity)
	return activity, nil
}

// DeleteActivity - deletes activity record with given ID from SQLite db.
func (s *SQLite) DeleteActivity(activityID string) (int64, error) {
	entry := s.logger.WithField("func", "DeleteActivity")

	entry.Debugf("Deleting activity with id: %s", activityID)
	result, err := s.Exec(activityDelete, activityID)

	if err != nil {
		return -1, fmt.Errorf("SQLite s.Exec(): %w", err)
	}

	id, err := result.RowsAffected()

	if err != nil {
		return -1, fmt.Errorf("SQLite LastInsertId(): %w", err)
	}

	entry.Debugf("Activity with id %s deleted successfully, rows affected: %d", activityID, id)
	return id, nil
}
