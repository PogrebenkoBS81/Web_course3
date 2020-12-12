package sqlite

import (
	"activity_api/common/models"
	"fmt"
)

// GetUserActivity - returns data about users activity between 2 dates (timestamps).
func (s *SQLite) GetUserActivity(userID, startTime, endTime string) (*models.UserActivity, error) {
	entry := s.logger.WithField("func", "GetUserActivity")
	entry.Debugf(
		"Retrieving user activity data, user id - %s, start time - %s, end time - %s",
		userID,
		startTime,
		endTime,
	)

	userActivity := new(models.UserActivity)
	query := s.buildActivityTimeQuery(getUsersActivity, startTime, endTime)

	if err := s.Pick(userActivity, query, userID); err != nil {
		return nil, fmt.Errorf("SQLite s.Pick(), activityGet: %w", err)
	}

	entry.Debugf("Retrieved user (id: %s) activity data: %+v", userID, userActivity)
	return userActivity, nil
}

// GetDepartmentActivity - returns data about users activity between 2 dates (timestamps).
func (s *SQLite) GetDepartmentActivity(departID, startTime, endTime string) (*models.DepartmentActivity, error) {
	entry := s.logger.WithField("func", "GetDepartmentActivity")
	entry.Debugf(
		"Retrieving department activity data, department id - %s, start time - %s, end time - %s",
		departID,
		startTime,
		endTime,
	)

	departmentActivity := new(models.DepartmentActivity)
	query := s.buildActivityTimeQuery(getDepartmentsActivity, startTime, endTime)

	if err := s.Pick(departmentActivity, query, departID); err != nil {
		return nil, fmt.Errorf("SQLite s.Pick(), activityGet: %w", err)
	}

	entry.Debugf("Retrieved department (id: %s) activity data: %+v", departID, departmentActivity)
	return departmentActivity, nil
}

// buildActivityTimeQuery - appends time check to query.
func (s *SQLite) buildActivityTimeQuery(query, timeBefore, timeAfter string) string {
	entry := s.logger.WithField("func", "buildActivityTimeQuery")
	entry.Debugf("Adding time check to query %s, start - %s, end - %s", query, timeBefore, timeAfter)

	if timeBefore != "" {
		// This sprintf is safe for SQL injection due to activityTimeStart enclosed in quotes ''
		query += fmt.Sprintf(activityTimeStart, timeBefore)
	}

	if timeAfter != "" {
		// Same as above
		query += fmt.Sprintf(activityTimeEnd, timeAfter)
	}

	entry.Debugf("Result query with time check: %s", query)
	return query
}
