package config

import (
	"github.com/rs/zerolog"
)

type EnvCache struct {
	RevPort    string `envconfig:"REVPORT" required:true`
	RouterPort string `envconfig:"ROUTERPORT" required:true`
	LogLevel   string `envconfig:"LOGLEVEL" required:true`
	Host       string `envconfig:"HOST" required:true`
	Port       string `envconfig:"PORT" required:true`
	User       string `envconfig:"USER" required:true`
	Password   string `envconfig:"PASSWORD" required:true`
	Dbname     string `envconfig:"DBNAME" required:true`
	Sslmode    string `envconfig:"SSLMODE" required:true`
}

// GetSSlmode returns field SSlMode
func (c EnvCache) GetSslmode() string {
	return c.Sslmode
}

// GetDbname returns field Dbname
func (c EnvCache) GetDbname() string {
	return c.Dbname
}

// GetPassword returns field Password
func (c EnvCache) GetPassword() string {
	return c.Password
}

// GetUser returns field User
func (c EnvCache) GetUser() string {
	return c.User
}

// GetPort returns field Port
func (c EnvCache) GetPort() string {
	return c.Port
}

// GetHost returns field Host
func (c EnvCache) GetHost() string {
	return c.Host
}

// GetRevPort returns field RevPort
func (c EnvCache) GetRevPort() string {
	return c.RevPort
}

// GetRouterPort returns field RouterPort
func (c EnvCache) GetRouterPort() string {
	return c.RouterPort
}

// GetLogLevel determines which level
// the LogLevel environment
// corresponds to.
func (c EnvCache) GetLogLevel() zerolog.Level {
	switch c.LogLevel {
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.Disabled
	}
}
