// Package config provides user configuration loading from ~/.config/linear/config.yaml
package config

import (
	"fmt"
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
}

// MCPConfig holds MCP-specific settings.
type MCPConfig struct {
	// FieldDefaults for MCP tools (overrides top-level FieldDefaults)
	FieldDefaults map[string]string `yaml:"field_defaults"`
}

// Load reads the user configuration file from the standard location.
// Returns nil if the file doesn't exist (not an error - config is optional).
func Load() (*Config, error) {
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
