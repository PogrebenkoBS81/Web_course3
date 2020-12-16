package control

import (
	"activity_api/common/key_generator"
	"activity_api/data_manager/cache"
	"fmt"
	"os"
)

const pingersNum = 2

// iManageable - interface for service control.
// If service is unavailable, pingers will try to recover service via Open\Close functions.
type iManageable interface {
	Describe() string
	Restart() error
	OK() error
}

// AAServiceConfig - config for AAService
type AAServiceConfig struct {
	CacheType int

	DbType     int    // db type, 0 - SQLite
	ConnString string // Conn string to DB
	Addr       string // Addr of service to listen

	LogLevel uint32 // Log level for logrus
	LogFile  string // File to log in // temporary unused, I didn't write log to a file for now

	Cache *cache.ICacheConfig // Config for cache manager
}

// Set keys to env of the project
func setKeys() error {
	private, public, err := key_generator.GenerateKey()

	if err != nil {
		return fmt.Errorf("generateKey(): %w", err)
	}
	// Keys sets only for service, and they exists only while service is alive.
	if err := os.Setenv("TOKEN_PRIVATE", private); err != nil {
		return fmt.Errorf("private os.Setenv(): %w", err)
	}

	if err := os.Setenv("TOKEN_PUBLIC", public); err != nil {
		return fmt.Errorf("public os.Setenv(): %w", err)
	}

	return nil
}

// init - inits keys before start.
func init() {
	if err := setKeys(); err != nil {
		panic(err)
	}
}
