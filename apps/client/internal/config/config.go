package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func getConfigPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user's config dir: %w", err)
	}

	configDir := filepath.Join(userConfigDir, "gophkeeper")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create config dir: %w", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

func readConfig(path string) (*AppConfig, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config AppConfig
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func writeConfig(path string, config *AppConfig) error {
	configFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AppConfig describes the configuration of the application, especially
// authorization/session details that are used for server requests.
type AppConfig struct {
	// AccessToken is current session's access token, issued by the server.
	AccessToken *string `json:"access_token"`
	// SessionExpiresAt is a timestamp when this access token will be invalid.
	SessionExpiresAt *time.Time `json:"session_expires_at"`
}

// LoadConfig loads local instance of the configuration file, performs
// validation and returns it.
//
// If session has been expired, it automatically will return an empty access
// token, prompting user to re-authorize.
func LoadConfig() (*AppConfig, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Reading config, if it doesn't exist - creating a new one.
	config, err := readConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			config = &AppConfig{}
			if err := writeConfig(configPath, config); err != nil {
				return nil, fmt.Errorf("failed to create config file: %w", err)
			}
			return config, nil
		}
		return nil, err
	}

	if config.SessionExpiresAt != nil && time.Now().After(*config.SessionExpiresAt) {
		config.AccessToken = nil
		config.SessionExpiresAt = nil
	}

	return config, nil
}

// WriteConfig writes given config into user's config path.
func WriteConfig(config *AppConfig) error {
	configPath, err := getConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}
	return writeConfig(configPath, config)
}
