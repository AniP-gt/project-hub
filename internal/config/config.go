package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName        = "project-hub"
	ConfigFileName = "project-hub.json"
)

type Config struct {
	DefaultProjectID     string `json:"defaultProjectID"`
	DefaultOwner         string `json:"defaultOwner"`
	DisableNotifications bool   `json:"disableNotifications"`
	DefaultItemLimit     int    `json:"defaultItemLimit"`
	DefaultExcludeDone   bool   `json:"defaultExcludeDone"`
}

// ResolvePath returns the canonical config file path using XDG Base Directory spec.
// Always uses ~/.config/project-hub/ regardless of OS.
func ResolvePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}

	appConfigDir := filepath.Join(homeDir, ".config", AppName)
	configPath := filepath.Join(appConfigDir, ConfigFileName)

	return configPath, nil
}

// Load reads configuration from a JSON file.
// Returns an empty Config{} (not an error) if the file doesn't exist.
// Returns a descriptive error if the file is malformed.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return cfg, nil
}

// Save writes configuration to a JSON file.
// Creates parent directories if they don't exist.
// Returns a descriptive error if the operation fails.
func Save(path string, cfg Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
