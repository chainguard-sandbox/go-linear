// Package config provides user configuration loading from ~/.config/linear/config.yaml
package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the user's configuration file structure.
type Config struct {
	// FieldDefaults maps command names to default field lists
	// Example: "issue.list": "id,identifier,title,state.name"
	FieldDefaults map[string]string `yaml:"field_defaults"`

	// MCP contains MCP-specific configuration
	MCP MCPConfig `yaml:"mcp"`

	// Defaults contains default values for commands
	Defaults DefaultsConfig `yaml:"defaults"`
}

// MCPConfig holds MCP-specific settings.
type MCPConfig struct {
	// FieldDefaults for MCP tools (overrides top-level FieldDefaults)
	FieldDefaults map[string]string `yaml:"field_defaults"`
}

// DefaultsConfig holds default values for command flags.
type DefaultsConfig struct {
	Team   string   `yaml:"team"`
	Labels []string `yaml:"labels"`
}

// Load reads configuration from both user and workspace locations.
// Workspace config (.linear-workspace.yaml) takes precedence over user config.
// Returns nil if no config files exist (not an error - config is optional).
func Load() (*Config, error) {
	// Load user config first
	userCfg, err := loadUserConfig()
	if err != nil {
		return nil, err
	}

	// Load workspace config (in current directory)
	workspaceCfg, err := loadWorkspaceConfig()
	if err != nil {
		return nil, err
	}

	// Merge: workspace overrides user
	return merge(userCfg, workspaceCfg), nil
}

// loadUserConfig reads the user configuration from standard locations.
func loadUserConfig() (*Config, error) {
	paths := []string{
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "linear", "config.yaml"),
		filepath.Join(os.Getenv("HOME"), ".config", "linear", "config.yaml"),
		filepath.Join(os.Getenv("HOME"), ".linear.yaml"), // Legacy
	}

	// Try each path
	for _, path := range paths {
		if path == "" {
			continue
		}

		cfg, err := loadFromPath(path)
		if err == nil {
			return cfg, nil
		}
		if !os.IsNotExist(err) {
			// Real error (not just missing file)
			return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
		}
	}

	// No config file found - return empty config (not an error)
	return &Config{}, nil
}

// loadWorkspaceConfig reads .linear-workspace.yaml from current directory.
func loadWorkspaceConfig() (*Config, error) {
	path := ".linear-workspace.yaml"
	cfg, err := loadFromPath(path)
	if os.IsNotExist(err) {
		return &Config{}, nil // No workspace config is fine
	}
	return cfg, err
}

// merge combines workspace config (higher priority) with user config (lower priority).
func merge(user, workspace *Config) *Config {
	// Handle nil inputs
	if user == nil {
		user = &Config{}
	}
	if workspace == nil {
		workspace = &Config{}
	}

	result := &Config{
		FieldDefaults: make(map[string]string),
	}

	// Copy field defaults: user first, then workspace overrides
	if user.FieldDefaults != nil {
		maps.Copy(result.FieldDefaults, user.FieldDefaults)
	}
	if workspace.FieldDefaults != nil {
		maps.Copy(result.FieldDefaults, workspace.FieldDefaults)
	}

	// MCP field defaults
	result.MCP.FieldDefaults = make(map[string]string)
	if user.MCP.FieldDefaults != nil {
		maps.Copy(result.MCP.FieldDefaults, user.MCP.FieldDefaults)
	}
	if workspace.MCP.FieldDefaults != nil {
		maps.Copy(result.MCP.FieldDefaults, workspace.MCP.FieldDefaults)
	}

	// Defaults: workspace overrides user
	if workspace.Defaults.Team != "" {
		result.Defaults.Team = workspace.Defaults.Team
	} else {
		result.Defaults.Team = user.Defaults.Team
	}

	if len(workspace.Defaults.Labels) > 0 {
		result.Defaults.Labels = workspace.Defaults.Labels
	} else {
		result.Defaults.Labels = user.Defaults.Labels
	}

	return result
}

// loadFromPath loads configuration from a specific file path.
func loadFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 - path from trusted config locations only
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	return &cfg, nil
}
