package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// AppName is the application name used in config directories.
	AppName = "project-hub"
	// ConfigFileName is the canonical JSON file name for projects configuration.
	ConfigFileName = "projects-tui.json"
)

// Config holds default settings for the projects-tui application.
type Config struct {
	// DefaultProjectID is the ID of the project to load on startup.
	DefaultProjectID string `json:"defaultProjectID"`
	// DefaultOwner is the GitHub owner to use for project queries.
	DefaultOwner string `json:"defaultOwner"`
	// DisableNotifications suppresses info-level notification messages in the UI.
	DisableNotifications bool `json:"disableNotifications"`
}

// Load reads configuration from a JSON file.
// Returns an empty Config{} (not an error) if the file doesn't exist.
// Returns a descriptive error if the file is malformed.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// File not found is not an error - return empty config
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		// Other errors (permissions, etc.) are reported
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
	// Create parent directories if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with indentation for readability
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ResolvePath returns the canonical config file path using os.UserConfigDir()
// for cross-platform compatibility.
//
// Returns path like: <UserConfigDir>/project-hub/projects-tui.json
// Example: ~/.config/project-hub/projects-tui.json (Linux)
//
//	~/Library/Application Support/project-hub/projects-tui.json (macOS)
//	%APPDATA%\project-hub\projects-tui.json (Windows)
func ResolvePath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user config directory: %w", err)
	}

	appConfigDir := filepath.Join(userConfigDir, AppName)
	configPath := filepath.Join(appConfigDir, ConfigFileName)

	return configPath, nil
}
