package config

import (
	"errors"
	"os"
)

// minJWTSecretLength keeps HS256 secrets at 256 bits or more.
const minJWTSecretLength = 32

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	CORSOrigin  string
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
		CORSOrigin:  os.Getenv("CORS_ORIGIN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if len(cfg.JWTSecret) < minJWTSecretLength {
		return Config{}, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if cfg.CORSOrigin == "" {
		return Config{}, errors.New("CORS_ORIGIN is required")
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
