package config

import "github.com/rs/zerolog"

type EnvCache struct {
	RevPort    string `envconfig:"REVPORT" required:true`
	RouterPort string `envconfig:"ROUTERPORT" required:true`
	LogLevel   string `envconfig:"LOGLEVEL" required:true`
}

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
