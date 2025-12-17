package fieldfilter

import (
	"testing"
)

func TestGetDefaults(t *testing.T) {
	// Test 1: Built-in defaults (no config)
	defaults := GetDefaults("issue.list", nil)
	if defaults == nil {
		t.Fatal("GetDefaults should return built-in defaults")
	}
	if len(defaults) != 8 {
		t.Errorf("issue.list defaults length = %d, want 8", len(defaults))
	}
	// Check first field
	if defaults[0] != "id" {
		t.Errorf("First default field = %q, want %q", defaults[0], "id")
	}

	// Test 2: Config override
	configOverrides := map[string]string{
		"issue.list": "id,title,url",
	}
	defaults = GetDefaults("issue.list", configOverrides)
	if len(defaults) != 3 {
		t.Errorf("Config override length = %d, want 3", len(defaults))
	}

	// Test 3: Unknown command
	defaults = GetDefaults("nonexistent.command", nil)
	if defaults != nil {
		t.Errorf("Unknown command should return nil, got %v", defaults)
	}

	// Test 4: Team defaults
	defaults = GetDefaults("team.list", nil)
	if len(defaults) != 7 {
		t.Errorf("team.list defaults length = %d, want 7 (includes issueCount)", len(defaults))
	}
}

func TestGetDefaultsForMCP(t *testing.T) {
	// Test 1: MCP-specific override
	mcpOverrides := map[string]string{
		"issue.list": "id,title",
	}
	configOverrides := map[string]string{
		"issue.list": "id,title,description",
	}

	defaults := GetDefaultsForMCP("issue.list", mcpOverrides, configOverrides)
	if len(defaults) != 2 {
		t.Errorf("MCP override should take precedence, got %d fields, want 2", len(defaults))
	}

	// Test 2: Fallback to config
	defaults = GetDefaultsForMCP("issue.list", nil, configOverrides)
	if len(defaults) != 3 {
		t.Errorf("Should fall back to config override, got %d fields, want 3", len(defaults))
	}

	// Test 3: Fallback to built-in
	defaults = GetDefaultsForMCP("issue.list", nil, nil)
	if len(defaults) != 8 {
		t.Errorf("Should fall back to built-in, got %d fields, want 8", len(defaults))
	}
}
