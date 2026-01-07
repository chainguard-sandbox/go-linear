package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create temp dir for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test 1: No config file (should return empty config, not error)
	_ = os.Setenv("HOME", tmpDir)
	_ = os.Setenv("XDG_CONFIG_HOME", "")
	defer os.Unsetenv("HOME")
	defer os.Unsetenv("XDG_CONFIG_HOME")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() with no config should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() should return empty config, not nil")
	}

	// Test 2: Valid config file
	configContent := `
field_defaults:
  issue.list: "id,title,state.name"
  team.get: "id,name,key"

mcp:
  field_defaults:
    issue.list: "id,title"
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".linear.yaml"), []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Check field defaults loaded
	if cfg.FieldDefaults == nil {
		t.Fatal("FieldDefaults should be populated")
	}
	if val, ok := cfg.FieldDefaults["issue.list"]; !ok || val != "id,title,state.name" {
		t.Errorf("FieldDefaults[issue.list] = %q, want %q", val, "id,title,state.name")
	}

	// Check MCP overrides
	if cfg.MCP.FieldDefaults == nil {
		t.Fatal("MCP.FieldDefaults should be populated")
	}
	if val, ok := cfg.MCP.FieldDefaults["issue.list"]; !ok || val != "id,title" {
		t.Errorf("MCP.FieldDefaults[issue.list] = %q, want %q", val, "id,title")
	}

	// Test 3: Invalid YAML (should error)
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0o600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Move to ~/.config/linear/config.yaml location
	linearDir := filepath.Join(tmpDir, ".config", "linear")
	if err := os.MkdirAll(linearDir, 0o750); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.Rename(configPath, filepath.Join(linearDir, "config.yaml")); err != nil {
		t.Fatalf("Failed to move config: %v", err)
	}

	_, err = Load()
	if err == nil {
		t.Error("Load() should error on invalid YAML")
	}
}

func TestLoadWorkspace(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	_ = os.Chdir(tmpDir)

	// Test 1: Workspace config only
	workspaceContent := `
defaults:
  team: Engineering
  labels:
    - bug
    - triage
`
	if err := os.WriteFile(".linear-workspace.yaml", []byte(workspaceContent), 0o600); err != nil {
		t.Fatalf("Failed to write workspace config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Defaults.Team != "Engineering" {
		t.Errorf("Workspace team = %q, want %q", cfg.Defaults.Team, "Engineering")
	}
	if len(cfg.Defaults.Labels) != 2 {
		t.Errorf("Workspace labels length = %d, want 2", len(cfg.Defaults.Labels))
	}

	// Test 2: Workspace overrides user
	_ = os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	userContent := `
defaults:
  team: Platform
  labels:
    - user-label
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".linear.yaml"), []byte(userContent), 0o600); err != nil {
		t.Fatalf("Failed to write user config: %v", err)
	}

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load() with both configs error: %v", err)
	}

	// Workspace should override user
	if cfg.Defaults.Team != "Engineering" {
		t.Errorf("Merged team = %q, want %q (workspace should override)", cfg.Defaults.Team, "Engineering")
	}
	if len(cfg.Defaults.Labels) != 2 || cfg.Defaults.Labels[0] != "bug" {
		t.Errorf("Merged labels = %v, want workspace labels", cfg.Defaults.Labels)
	}
}
