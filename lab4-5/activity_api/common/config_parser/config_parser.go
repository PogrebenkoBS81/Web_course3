package config_parser

import (
	"activity_api/control"
	"activity_api/data_manager/db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// parseJSON - parses json by given path to data.
func parseJSON(filePath string, data interface{}) error {
	file, err := os.Open(filePath)

	if err != nil {
		return fmt.Errorf("os.Open(): %w", err)
	}
	// Wrap in lambda function is required, due to
	defer func() {
		// Since i don't want to pass logrus here,
		// but i don't want to miss an error during config read,
		// panic is used.
		if err := file.Close(); err != nil {
			panic(err)
		}
	}() // Check missing error

	bytes, err := ioutil.ReadAll(file)

	if err != nil {
		return fmt.Errorf("ioutil.ReadAll(): %w", err)
	}

	if err = json.Unmarshal(bytes, data); err != nil {
		return fmt.Errorf("json.Unmarshal(): %w", err)
	}

	return nil
}

// Check if path is abs, if not - make it abs.
// Required if app will run under system.
func EnsureAbsPath(toCheck string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if err != nil {
		return "", fmt.Errorf("os.Getwd(): %w", err)
	}

	return filepath.Join(dir, toCheck), nil
}

// ParseConfig - parses AAServer config by given path.
func ParseConfig(configName string) (*control.AAServiceConfig, error) {
	// Make path abs
	path, err := EnsureAbsPath(configName)

	if err != nil {
		return nil, fmt.Errorf("EnsureAbsPath(): %w", err)
	}

	config := new(control.AAServiceConfig)
	// Parse config
	if err = parseJSON(path, config); err != nil {
		return nil, fmt.Errorf("parseJSON(): %w", err)
	}
	// Due to SQLite conn string is DB path - ensure that this path is abs
	if config.DbType == db.SQLite {
		conn, err := EnsureAbsPath(config.ConnString)

		if err != nil {
			return nil, fmt.Errorf("EnsureAbsPath(): %w", err)
		}

		config.ConnString = conn
	}

	return config, nil
}
