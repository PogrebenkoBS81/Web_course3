package core

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"sync"
)

// ISQLCore - common interface for all SQL databases.
// With it, a lot of copy-paste code would be avoided during integration of new SQL DB,
// like MySQL, MSSQL, etc
type ISQLCore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Pick(dest interface{}, query string, args ...interface{}) error

	Open() error
	Close() error
	Restart() error
	OK() error
}

// SQL core struct
type SQL struct {
	driver     string       // Driver of given SQL DB
	connString string       // Conn string of given SQL DB
	mtx        sync.RWMutex // RWMutex to improve performance
	db         *sqlx.DB     // Connection to DB

	logger logrus.FieldLogger
}

// NewSQL - returns new SQL core struct
func NewSQL(driver, connString string, logger logrus.FieldLogger) *SQL {
	return &SQL{
		driver:     driver,
		connString: connString,
		logger:     logger.WithField("module", "SQLCore"),
	}
}

// Restart - restarts the SQL db, can be user when service is down.
func (s *SQL) Restart() (err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	entry := s.logger.WithField("func", "Restart")
	entry.Info("Restarting SQL database...")

	if err = s.close(); err != nil {
		entry.Errorf("close redis connection error")
	}

	if err = s.open(); err != nil {
		entry.Errorf("close redis connection error")
	}

	return
}

// Open - opens connection with given SQL db
func (s *SQL) Open() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.open()
}

// open - open helper. Will be user in Restart and in Open.
func (s *SQL) open() (err error) {
	s.logger.WithField("func", "open").Info("Opening DB connection...")

	if s.db != nil {
		return errors.New("sql connection already exist")
	}

	s.db, err = sqlx.Open(s.driver, s.connString)
	if err != nil {
		return fmt.Errorf("SQL Open(): %w", err)
	}

	return
}

// Close - closes connection with given DB
func (s *SQL) Close() (err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	return s.close()
}

// close - close helper. Will be user in Restart and in Close.
func (s *SQL) close() (err error) {
	s.logger.WithField("func", "close").Info("Closing DB connection...")

	if s.db == nil {
		return errors.New("sql connection doesn't exist")
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("SQL Close(): %w", err)
	}

	s.db = nil // Some DB implementations doesn't require it, but I'b better be sure.

	return
}

// OK - pings DB to check it status
func (s *SQL) OK() (err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.db == nil {
		return errors.New("sql connection doesn't exist")
	}

	s.logger.WithField("func", "OK").Debug("Doing ping...")

	// TODO: if ping will be used not only in pinger - create flag to reduce overhead.
	return s.db.Ping()
}

// Exec - runs given query on db
func (s *SQL) Exec(query string, args ...interface{}) (sql.Result, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.db == nil {
		return nil, errors.New("sql connection doesn't exist")
	}

	s.logger.WithField("func", "Exec").Debugf("Executing query: %s", query)

	result, err := s.db.Exec(query, args...)

	if err != nil {
		return nil, fmt.Errorf("SQL Exec(): %w", err)
	}

	return result, nil
}

// Get - writes result of query to given interface.
func (s *SQL) Get(dest interface{}, query string, args ...interface{}) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.db == nil {
		return errors.New("sql connection doesn't exist")
	}

	s.logger.WithField("func", "Get").Debugf("Get query: %s", query)

	if err := s.db.Select(dest, query, args...); err != nil {
		return fmt.Errorf("SQL conn.Select(): %w", err)
	}

	return nil
}

// Get - writes single object from sql query to given interface.
func (s *SQL) Pick(dest interface{}, query string, args ...interface{}) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if s.db == nil {
		return errors.New("sql connection doesn't exist")
	}

	s.logger.WithField("func", "Pick").Debugf("Pick query: %s", query)

	if err := s.db.Get(dest, query, args...); err != nil {
		return fmt.Errorf("SQL conn.Get(): %w", err)
	}

	return nil
}
