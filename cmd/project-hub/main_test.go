package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"project-hub/internal/config"
)

func TestResolveStartupOptionsCLIWins(t *testing.T) {
	tests := []struct {
		name        string
		cliProject  string
		cliOwner    string
		cfg         config.Config
		wantProject string
		wantOwner   string
	}{
		{
			name:        "CLI project and owner override config",
			cliProject:  "cli-project-id",
			cliOwner:    "cli-owner",
			cfg:         config.Config{DefaultProjectID: "config-project-id", DefaultOwner: "config-owner"},
			wantProject: "cli-project-id",
			wantOwner:   "cli-owner",
		},
		{
			name:        "CLI project overrides config project",
			cliProject:  "cli-project-id",
			cliOwner:    "",
			cfg:         config.Config{DefaultProjectID: "config-project-id", DefaultOwner: "config-owner"},
			wantProject: "cli-project-id",
			wantOwner:   "config-owner",
		},
		{
			name:        "CLI owner overrides config owner",
			cliProject:  "",
			cliOwner:    "cli-owner",
			cfg:         config.Config{DefaultProjectID: "config-project-id", DefaultOwner: "config-owner"},
			wantProject: "config-project-id",
			wantOwner:   "cli-owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, owner := resolveStartupOptions(tt.cliProject, tt.cliOwner, tt.cfg)
			if project != tt.wantProject {
				t.Errorf("project: got %q, want %q", project, tt.wantProject)
			}
			if owner != tt.wantOwner {
				t.Errorf("owner: got %q, want %q", owner, tt.wantOwner)
			}
		})
	}
}

func TestResolveStartupOptionsFallback(t *testing.T) {
	tests := []struct {
		name        string
		cliProject  string
		cliOwner    string
		cfg         config.Config
		wantProject string
		wantOwner   string
	}{
		{
			name:        "Both CLI empty, use config values",
			cliProject:  "",
			cliOwner:    "",
			cfg:         config.Config{DefaultProjectID: "config-project-id", DefaultOwner: "config-owner"},
			wantProject: "config-project-id",
			wantOwner:   "config-owner",
		},
		{
			name:        "All empty values",
			cliProject:  "",
			cliOwner:    "",
			cfg:         config.Config{DefaultProjectID: "", DefaultOwner: ""},
			wantProject: "",
			wantOwner:   "",
		},
		{
			name:        "Only config project set",
			cliProject:  "",
			cliOwner:    "",
			cfg:         config.Config{DefaultProjectID: "config-project-id", DefaultOwner: ""},
			wantProject: "config-project-id",
			wantOwner:   "",
		},
		{
			name:        "Only config owner set",
			cliProject:  "",
			cliOwner:    "",
			cfg:         config.Config{DefaultProjectID: "", DefaultOwner: "config-owner"},
			wantProject: "",
			wantOwner:   "config-owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, owner := resolveStartupOptions(tt.cliProject, tt.cliOwner, tt.cfg)
			if project != tt.wantProject {
				t.Errorf("project: got %q, want %q", project, tt.wantProject)
			}
			if owner != tt.wantOwner {
				t.Errorf("owner: got %q, want %q", owner, tt.wantOwner)
			}
		})
	}
}

func TestLoadStartupConfigUsesSavedDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	configPath, err := config.ResolvePath()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"defaultProjectID":"9","defaultOwner":"AniP-gt"}`), 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	var stderr bytes.Buffer
	cfg, _ := loadStartupConfig(&stderr)

	if cfg.DefaultProjectID != "9" {
		t.Fatalf("expected project id 9, got %q", cfg.DefaultProjectID)
	}
	if cfg.DefaultOwner != "AniP-gt" {
		t.Fatalf("expected owner AniP-gt, got %q", cfg.DefaultOwner)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected no warning output, got %q", stderr.String())
	}
}

func TestLoadStartupConfigMalformedConfigWarning(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	configPath, err := config.ResolvePath()
	if err != nil {
		t.Fatalf("resolve config path: %v", err)
	}
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`{"defaultProjectID":"9",`), 0o644); err != nil {
		t.Fatalf("write malformed config file: %v", err)
	}

	var stderr bytes.Buffer
	cfg, _ := loadStartupConfig(&stderr)

	if cfg.DefaultProjectID != "" || cfg.DefaultOwner != "" {
		t.Fatalf("expected empty fallback config, got %+v", cfg)
	}
	if !strings.Contains(stderr.String(), "warning: failed to load config") {
		t.Fatalf("expected malformed-config warning, got %q", stderr.String())
	}
}
