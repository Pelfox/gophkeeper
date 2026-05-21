package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// AppConfig holds application's configuration.
type AppConfig struct {
	// ListenAddr is the address for the HTTP server to listen on.
	ListenAddr string `koanf:"listen_addr"`
	// DatabaseDSN contains the URL that is used to connect to PostgreSQL.
	DatabaseDSN string `koanf:"database_dsn"`
	// JWTSecret is the secret for JWT signing process.
	JWTSecret string `koanf:"jwt_secret"`
}

// LoadConfig tries to load configuration for the application, using
// environment variables.
func LoadConfig() (*AppConfig, error) {
	_ = godotenv.Load()

	var configParser = koanf.New(".")
	cfg := AppConfig{ListenAddr: ":3000"}

	err := configParser.Load(env.Provider(".", env.Opt{
		Prefix: "GOPHKEEPER_",
		TransformFunc: func(key string, value string) (string, any) {
			key = strings.ToLower(strings.TrimPrefix(key, "GOPHKEEPER_"))
			if strings.Contains(value, " ") {
				return key, strings.Split(value, " ")
			}
			return key, value
		},
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	if err := configParser.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return &cfg, nil
}
