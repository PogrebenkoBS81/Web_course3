package cli

import (
	"flag"
)

// Config - config with server host/port
type Config struct {
	Host string
	Port int
	Interval int
}

var (
	fHost = flag.String(
		"host",
		"localhost",
		"Server host.",
	)

	fPort = flag.Int(
		"port",
		8080,
		"Server port.",
	)

	fInterval = flag.Int(
		"interval",
		23,
		"Connection type.",
	)
)

// ParseArgs - parses cmd arguments and returns config with data
func ParseArgs() *Config {
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Config{
		Host: *fHost,
		Port: *fPort,
		Interval: *fInterval,
	}
}

