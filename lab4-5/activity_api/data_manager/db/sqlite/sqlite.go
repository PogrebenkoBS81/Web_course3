package sqlite

import (
	"activity_api/data_manager/db/core"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

const serviceName = "SQLite"

// SQLite - sqlite service
type SQLite struct {
	core.ISQLCore
	logger logrus.FieldLogger
}

// NewSQLite - returns new SQLite DB
func NewSQLite(connString string, logger logrus.FieldLogger) core.ISQLDatabase {
	return &SQLite{
		ISQLCore: core.NewSQL("sqlite3", connString, logger),
		logger:   logger.WithField("module", "SQLite"),
	}
}

// CreateDB - creates required tables for AAService (CREATE IF NOT EXIST)
func (s *SQLite) CreateDB() error {
	entry := s.logger.WithField("func", "CreateDB")

	entry.Info("Initializing admins table...")
	_, err := s.Exec(createAdminsTable)

	if err != nil {
		return fmt.Errorf("CreateDB(), createAdminsTable: %w", err)
	}

	entry.Info("Initializing departments table")
	_, err = s.Exec(createDepartmentsTable)

	if err != nil {
		return fmt.Errorf("CreateDB(), createDepartmentsTable: %w", err)
	}

	entry.Info("Initializing users table")
	_, err = s.Exec(createUsersTable)

	if err != nil {
		return fmt.Errorf("CreateDB(), createUsersTable: %w", err)
	}

	entry.Info("Initializing activity table")
	_, err = s.Exec(createActivityTable)

	if err != nil {
		return fmt.Errorf("CreateDB(), createActivityTable: %w", err)
	}

	return nil
}

// Describe - returns SQLite name, so it's possible to identify service behind interface.
func (s *SQLite) Describe() string {
	s.logger.WithField("func", "Describe").Debug("Getting SQLite description")

	return serviceName
}
