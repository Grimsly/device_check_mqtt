package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	DevicesCSVFilepath string
}

// Loads the config from flags
//
// --port Port number
//
// --devices_path Path to devices CSV file
func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", "6733", "Server port")
	flag.StringVar(&cfg.DevicesCSVFilepath, "devices_path", "data/devices.csv", "Path to devices CSV file")
	flag.Parse()

	// Checks to make sure the user entered a valid port number
	if cfg.Port == "" {
		return nil, fmt.Errorf("Port is required")
	}
	port, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("Port entered is not a number")
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("Port must be between 1 and 65535")
	}

	// Check if the filepath provided for the devices file is valid
	if _, err := os.Stat(cfg.DevicesCSVFilepath); err != nil {
		return nil, fmt.Errorf("Cannot access devices CSV file: %w", err)
	}
	return cfg, nil
}
