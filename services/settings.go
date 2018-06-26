package services

import (
	"time"

	"github.com/caarlos0/env"
)

// LastUpdatedTime is the last time the system updated the status db
var LastUpdatedTime time.Time

// SystemSettings are the required settings for the system to know how to run
type SystemSettings struct {
	Port         string `env:"HTTP_PORT"`
	Mode         string `env:"GIN_MODE"`
	UpdateTimer  int    `env:"UPDATE_TIME"`
	DoorbotToken string `env:"DOORBOT_API_KEY"`
}

// GetSystemSettings will return the current settings parsed by environment variables
func GetSystemSettings() SystemSettings {
	var settings = SystemSettings{}
	env.Parse(&settings)

	return settings
}
