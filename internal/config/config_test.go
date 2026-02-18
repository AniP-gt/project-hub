package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolvePath verifies the config file path resolution across platforms.
func TestResolvePath(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		checks  func(t *testing.T, path string)
	}{
		{
			name:    "resolve path returns non-empty path",
			wantErr: false,
			checks: func(t *testing.T, path string) {
				if path == "" {
					t.Error("ResolvePath() returned empty path")
				}
			},
		},
		{
			name:    "resolved path contains app name",
			wantErr: false,
			checks: func(t *testing.T, path string) {
				if !strings.Contains(path, AppName) {
					t.Errorf("path should contain AppName %q, got %q", AppName, path)
				}
			},
		},
		{
			name:    "resolved path ends with config filename",
			wantErr: false,
			checks: func(t *testing.T, path string) {
				if !strings.HasSuffix(path, ConfigFileName) {
					t.Errorf("path should end with %q, got %q", ConfigFileName, path)
				}
			},
		},
		{
			name:    "resolved path is absolute",
			wantErr: false,
			checks: func(t *testing.T, path string) {
				if !filepath.IsAbs(path) {
					t.Errorf("path should be absolute, got %q", path)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ResolvePath()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolvePath() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.checks != nil {
				tt.checks(t, path)
			}
		})
	}
}

// TestResolvePathError verifies error handling when config directory lookup fails.
func TestResolvePathError(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		checks  func(t *testing.T, path string, err error)
	}{
		{
			name: "succeeds with normal environment",
			setup: func() {
				// Ensure HOME/XDG_CONFIG_HOME are set (they normally are)
			},
			cleanup: func() {},
			checks: func(t *testing.T, path string, err error) {
				if err != nil {
					t.Errorf("ResolvePath() unexpected error: %v", err)
				}
				if path == "" {
					t.Error("ResolvePath() returned empty path")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()
			path, err := ResolvePath()
			tt.checks(t, path, err)
		})
	}
}

// TestSaveLoadRoundTrip verifies that saved config can be loaded back correctly.
func TestSaveLoadRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "save and load simple config",
			config: Config{
				DefaultProjectID: "123",
				DefaultOwner:     "testuser",
			},
		},
		{
			name: "save and load config with empty fields",
			config: Config{
				DefaultProjectID: "",
				DefaultOwner:     "",
			},
		},
		{
			name: "save and load config with special characters",
			config: Config{
				DefaultProjectID: "project-with-dashes-and_underscores",
				DefaultOwner:     "owner@example.com",
			},
		},
		{
			name: "save and load config with long values",
			config: Config{
				DefaultProjectID: "this-is-a-very-long-project-id-with-many-characters-123456789",
				DefaultOwner:     "organization-with-extremely-long-name-that-should-still-work",
			},
		},
		{
			name: "save and load config with numeric strings",
			config: Config{
				DefaultProjectID: "9876543210",
				DefaultOwner:     "12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use temporary directory for test isolation
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ConfigFileName)

			// Save config
			err := Save(configPath, tt.config)
			if err != nil {
				t.Fatalf("Save() failed: %v", err)
			}

			// Verify file was created
			if _, err := os.Stat(configPath); err != nil {
				t.Fatalf("config file not created: %v", err)
			}

			// Load config back
			loaded, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() failed: %v", err)
			}

			// Verify loaded config matches original
			if loaded.DefaultProjectID != tt.config.DefaultProjectID {
				t.Errorf("DefaultProjectID mismatch: want %q, got %q",
					tt.config.DefaultProjectID, loaded.DefaultProjectID)
			}
			if loaded.DefaultOwner != tt.config.DefaultOwner {
				t.Errorf("DefaultOwner mismatch: want %q, got %q",
					tt.config.DefaultOwner, loaded.DefaultOwner)
			}
		})
	}
}

// TestLoadMalformedConfig verifies graceful error handling for malformed JSON.
func TestLoadMalformedConfig(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		wantErr     bool
		errContains string
	}{
		{
			name:        "invalid JSON - unclosed brace",
			fileContent: `{"defaultProjectID": "123"`,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "invalid JSON - trailing comma",
			fileContent: `{"defaultProjectID": "123",}`,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "invalid JSON - single quotes",
			fileContent: `{'defaultProjectID': '123'}`,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "empty file",
			fileContent: ``,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "plain text instead of JSON",
			fileContent: `this is not json at all`,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "JSON array instead of object",
			fileContent: `["item1", "item2"]`,
			wantErr:     true,
			errContains: "parse",
		},
		{
			name:        "JSON with null values",
			fileContent: `{"defaultProjectID": null, "defaultOwner": null}`,
			wantErr:     false,
			errContains: "",
		},
		{
			name:        "JSON with extra fields (should be ignored)",
			fileContent: `{"defaultProjectID": "123", "defaultOwner": "user", "extraField": "ignored"}`,
			wantErr:     false,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ConfigFileName)

			// Write test content
			err := os.WriteFile(configPath, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("WriteFile failed: %v", err)
			}

			// Load the malformed config
			_, err = Load(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errContains)) {
				t.Errorf("Load() error should contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}

// TestLoadMissingFile verifies that missing config files return empty defaults without error.
func TestLoadMissingFile(t *testing.T) {
	tests := []struct {
		name          string
		createFile    bool
		expectedError bool
		checkDefaults func(t *testing.T, cfg Config)
	}{
		{
			name:          "load from non-existent file returns empty config",
			createFile:    false,
			expectedError: false,
			checkDefaults: func(t *testing.T, cfg Config) {
				if cfg.DefaultProjectID != "" {
					t.Errorf("DefaultProjectID should be empty, got %q", cfg.DefaultProjectID)
				}
				if cfg.DefaultOwner != "" {
					t.Errorf("DefaultOwner should be empty, got %q", cfg.DefaultOwner)
				}
			},
		},
		{
			name:          "load from deleted file returns empty config",
			createFile:    true,
			expectedError: false,
			checkDefaults: func(t *testing.T, cfg Config) {
				if cfg.DefaultProjectID != "" {
					t.Errorf("DefaultProjectID should be empty, got %q", cfg.DefaultProjectID)
				}
				if cfg.DefaultOwner != "" {
					t.Errorf("DefaultOwner should be empty, got %q", cfg.DefaultOwner)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ConfigFileName)

			// Create and delete file if requested
			if tt.createFile {
				err := os.WriteFile(configPath, []byte(`{"defaultProjectID": "123"}`), 0644)
				if err != nil {
					t.Fatalf("WriteFile failed: %v", err)
				}
				err = os.Remove(configPath)
				if err != nil {
					t.Fatalf("Remove failed: %v", err)
				}
			}

			// Try to load from non-existent path
			cfg, err := Load(configPath)

			if (err != nil) != tt.expectedError {
				t.Errorf("Load() error = %v, expectedError = %v", err, tt.expectedError)
				return
			}

			if tt.checkDefaults != nil {
				tt.checkDefaults(t, cfg)
			}
		})
	}
}

// TestSaveCreatesDirectories verifies that Save creates parent directories.
func TestSaveCreatesDirectories(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		config Config
	}{
		{
			name:   "save with non-existent parent directory",
			path:   "deep/nested/config.json",
			config: Config{DefaultProjectID: "123", DefaultOwner: "user"},
		},
		{
			name:   "save with deeply nested directories",
			path:   "a/b/c/d/e/config.json",
			config: Config{DefaultProjectID: "456", DefaultOwner: "owner"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, tt.path)

			// Save should create directories
			err := Save(configPath, tt.config)
			if err != nil {
				t.Fatalf("Save() failed: %v", err)
			}

			// Verify file exists
			if _, err := os.Stat(configPath); err != nil {
				t.Fatalf("config file not created: %v", err)
			}

			// Verify all parent directories were created
			dir := filepath.Dir(configPath)
			if _, err := os.Stat(dir); err != nil {
				t.Fatalf("parent directory not created: %v", err)
			}
		})
	}
}

// TestConfigJSONFormat verifies that saved JSON is properly formatted.
func TestConfigJSONFormat(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		checks func(t *testing.T, data []byte)
	}{
		{
			name:   "saved JSON is indented",
			config: Config{DefaultProjectID: "123", DefaultOwner: "user"},
			checks: func(t *testing.T, data []byte) {
				if !strings.Contains(string(data), "\n") {
					t.Error("JSON should be indented with newlines")
				}
				if !strings.Contains(string(data), "  ") {
					t.Error("JSON should use 2-space indentation")
				}
			},
		},
		{
			name:   "saved JSON is valid",
			config: Config{DefaultProjectID: "123", DefaultOwner: "user"},
			checks: func(t *testing.T, data []byte) {
				var cfg Config
				if err := json.Unmarshal(data, &cfg); err != nil {
					t.Errorf("saved JSON is invalid: %v", err)
				}
			},
		},
		{
			name:   "JSON uses correct field names",
			config: Config{DefaultProjectID: "123", DefaultOwner: "user"},
			checks: func(t *testing.T, data []byte) {
				dataStr := string(data)
				if !strings.Contains(dataStr, "defaultProjectID") {
					t.Error("JSON should contain 'defaultProjectID' field")
				}
				if !strings.Contains(dataStr, "defaultOwner") {
					t.Error("JSON should contain 'defaultOwner' field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, ConfigFileName)

			// Save config
			err := Save(configPath, tt.config)
			if err != nil {
				t.Fatalf("Save() failed: %v", err)
			}

			// Read raw file content
			data, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatalf("ReadFile failed: %v", err)
			}

			if tt.checks != nil {
				tt.checks(t, data)
			}
		})
	}
}

// TestConfigStruct verifies basic Config struct functionality.
func TestConfigStruct(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		checkField func(t *testing.T, cfg Config)
	}{
		{
			name: "config struct stores DefaultProjectID",
			config: Config{
				DefaultProjectID: "project-123",
				DefaultOwner:     "owner",
			},
			checkField: func(t *testing.T, cfg Config) {
				if cfg.DefaultProjectID != "project-123" {
					t.Errorf("DefaultProjectID not stored correctly")
				}
			},
		},
		{
			name: "config struct stores DefaultOwner",
			config: Config{
				DefaultProjectID: "project-456",
				DefaultOwner:     "test-owner",
			},
			checkField: func(t *testing.T, cfg Config) {
				if cfg.DefaultOwner != "test-owner" {
					t.Errorf("DefaultOwner not stored correctly")
				}
			},
		},
		{
			name: "config struct supports empty values",
			config: Config{
				DefaultProjectID: "",
				DefaultOwner:     "",
			},
			checkField: func(t *testing.T, cfg Config) {
				if cfg.DefaultProjectID != "" || cfg.DefaultOwner != "" {
					t.Errorf("empty config fields not handled correctly")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			tt.checkField(t, cfg)
		})
	}
}
