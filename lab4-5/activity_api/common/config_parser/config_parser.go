package config_parser

import (
	"activity_api/control"
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

// ParseConfig - parses AAServer config by given path.
func ParseConfig(configName string) (*control.AAServiceConfig, error) {
	// Get full path. Required if service will run under system.
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if err != nil {
		return nil, fmt.Errorf("os.Getwd(): %w", err)
	}

	configPath := filepath.Join(dir, configName)
	config := new(control.AAServiceConfig)

	if err = parseJSON(configPath, config); err != nil {
		return nil, fmt.Errorf("parseJSON(): %w", err)
	}

	//// Check if path is abs, if not - make it abs.
	//if abs := filepath.IsAbs(config.LogFile); !abs {
	//	pth, err := filepath.Abs(config.LogFile)
	//
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	config.LogFile = filepath.Join(dir, pth)
	//}

	return config, nil
}
