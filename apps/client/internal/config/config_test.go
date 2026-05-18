package config

import (
	"os"
	"testing"
	"time"
)

// TestLoadConfigCreatesMissingConfig verifies that a missing config file is
// created with empty session details.
func TestLoadConfigCreatesMissingConfig(t *testing.T) {
	setConfigHome(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AccessToken != nil {
		t.Fatalf("expected access token to be empty")
	}
	if cfg.SessionExpiresAt != nil {
		t.Fatalf("expected session expiration to be empty")
	}

	configPath, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath returned error: %v", err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}

// TestLoadConfigKeepsValidSession verifies that a non-expired session remains
// available after loading config.
func TestLoadConfigKeepsValidSession(t *testing.T) {
	setConfigHome(t)
	token := "access-token"
	expiresAt := time.Now().Add(time.Hour)

	if err := WriteConfig(&AppConfig{
		AccessToken:      &token,
		SessionExpiresAt: &expiresAt,
	}); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AccessToken == nil || *cfg.AccessToken != token {
		t.Fatalf("expected access token to be preserved")
	}
	if cfg.SessionExpiresAt == nil || !cfg.SessionExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected session expiration to be preserved")
	}
}

// TestLoadConfigClearsExpiredSession verifies that expired session details are
// cleared when config is loaded.
func TestLoadConfigClearsExpiredSession(t *testing.T) {
	setConfigHome(t)
	token := "access-token"
	expiresAt := time.Now().Add(-time.Hour)

	if err := WriteConfig(&AppConfig{
		AccessToken:      &token,
		SessionExpiresAt: &expiresAt,
	}); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AccessToken != nil {
		t.Fatalf("expected expired access token to be cleared")
	}
	if cfg.SessionExpiresAt != nil {
		t.Fatalf("expected expired session time to be cleared")
	}
}

// TestWriteConfigWritesJSON verifies that config values can be written and
// read back through the public config API.
func TestWriteConfigWritesJSON(t *testing.T) {
	setConfigHome(t)
	token := "access-token"
	expiresAt := time.Now().Add(time.Hour).UTC().Round(0)

	if err := WriteConfig(&AppConfig{
		AccessToken:      &token,
		SessionExpiresAt: &expiresAt,
	}); err != nil {
		t.Fatalf("WriteConfig returned error: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.AccessToken == nil || *cfg.AccessToken != token {
		t.Fatalf("unexpected access token: %#v", cfg.AccessToken)
	}
	if cfg.SessionExpiresAt == nil || !cfg.SessionExpiresAt.Equal(expiresAt) {
		t.Fatalf("unexpected session expiration: %#v", cfg.SessionExpiresAt)
	}
}

func setConfigHome(t *testing.T) string {
	t.Helper()

	configRoot := t.TempDir()
	t.Setenv("HOME", configRoot)
	t.Setenv("XDG_CONFIG_HOME", configRoot)

	return configRoot
}
