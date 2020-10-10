package cli

import (
	"flag"
)

// Config - config with server host/port
type Config struct {
	Host string
	Port int
	Clients int
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

	fClients = flag.Int(
		"clients",
		13,
		"Number of clients.",
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
		Clients: *fClients,
	}
}


