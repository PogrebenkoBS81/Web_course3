package main

import (
	"activity_api/common/config_parser"
	"activity_api/control"
)

func main() {
	// Get config from root directory
	config, err := config_parser.ParseConfig("config.json")

	if err != nil {
		panic(err)
	}

	// Run AAService
	control.NewAAService(config).Run()
}
