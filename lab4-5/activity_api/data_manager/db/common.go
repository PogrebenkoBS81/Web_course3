package db

import (
	"activity_api/data_manager/db/core"
	"activity_api/data_manager/db/sqlite"
	"github.com/sirupsen/logrus"
)

// All possible DB types. Service could easily migrate to another SQL db.
const (
	SQLite = iota
	// MSSql
	// MySql
	// ...
)

// NewAADatabase - returns new AAService database interface.
func NewAADatabase(dbType int, connString string, logger logrus.FieldLogger) core.ISQLDatabase {
	switch dbType {
	case SQLite:
		return sqlite.NewSQLite(connString, logger)
	//case MSSql:
	//	...
	default:
		logger.WithField("func", "NewAADatabase").
			Warnf("Unsupported dbType: %d, using default SQLite", dbType)

		return sqlite.NewSQLite(connString, logger)
	}
}
