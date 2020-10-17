package config

import (
	"github.com/kelseyhightower/envconfig"
)

// NewEnvSettings returns EnvSettings initialized from environment variables.
// `prefix` allows to add an extra prefix that needs to be used with all env var names.
func NewEnvSettings(prefix string) (EnvSettings, error) {
	var s EnvSettings
	return s, envconfig.Process(prefix, &s)
}

// EnvSettings reads settings from environment variables.
type EnvSettings struct {
	EnvHTTPListenPort  int    `envconfig:"HTTP_PORT" default:"8080"`
	EnvLogLevel        string `envconfig:"LOG_LEVEL" default:"info"`
	EnvStorageFilePath string `envconfig:"STORAGE_FILEPATH" default:"jobtome.dat"`
}

// HTTPPort returns a port number to listening for incoming HTTP connections.
func (es EnvSettings) HTTPPort() int {
	return es.EnvHTTPListenPort
}

// LogLevel returns a logging level.
func (es EnvSettings) LogLevel() string {
	return es.EnvLogLevel
}

// StorageFilePath returns path to the file to use as persisted storage.
func (es EnvSettings) StorageFilePath() string {
	return es.EnvStorageFilePath
}
