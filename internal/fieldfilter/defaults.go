// Package fieldfilter provides default field sets for commands.
//
// Defaults are versioned and considered part of the stable API.
// Fields may be added in minor versions (non-breaking), but not removed.
package fieldfilter

import "strings"

// BuiltinDefaults defines the default field sets for all commands.
// These are conservative selections covering 80% of use cases while
// allowing room for growth in future releases without breaking changes.
var BuiltinDefaults = map[string][]string{
	// Issue commands
	"issue.list": {
		"id", "identifier", "title", "url",
		"state.name", "team.key", "priority", "createdAt",
	},
	"issue.get": {
		"id", "identifier", "title", "url",
		"state.name", "team.key", "priority", "createdAt",
		"description", "assignee.name",
	},

	// Team commands
	"team.list": {
		"id", "name", "key", "description",
		"icon", "createdAt", "issueCount",
	},
	"team.get": {
		"id", "name", "key", "description",
		"icon", "createdAt", "color", "private", "issueCount",
	},

	// User commands
	"user.list": {
		"id", "name", "displayName", "email",
		"active", "avatarUrl",
	},
	"user.get": {
		"id", "name", "displayName", "email",
		"active", "avatarUrl", "admin",
	},

	// Comment commands
	"comment.list": {
		"id", "body", "createdAt", "user.name", "url",
	},
	"comment.get": {
		"id", "body", "createdAt", "user.name", "url",
		"editedAt",
	},

	// Label commands
	"label.list": {
		"id", "name", "color", "createdAt",
	},
	"label.get": {
		"id", "name", "color", "createdAt",
		"description",
	},

	// Workflow state commands
	"state.list": {
		"id", "name", "type", "color", "position",
	},
	"state.get": {
		"id", "name", "type", "color", "position",
	},

	// Cycle commands
	"cycle.list": {
		"id", "name", "startsAt", "endsAt", "createdAt",
	},
	"cycle.get": {
		"id", "name", "startsAt", "endsAt", "createdAt",
		"description",
	},

	// Attachment commands
	"attachment.list": {
		"id", "title", "url", "source", "createdAt",
	},
	"attachment.get": {
		"id", "title", "url", "source", "createdAt",
	},

	// Project commands
	"project.list": {
		"id", "name", "description", "createdAt",
	},
	"project.get": {
		"id", "name", "description", "createdAt",
		"color", "state",
	},

	// Roadmap commands
	"roadmap.list": {
		"id", "name", "description", "createdAt",
	},
	"roadmap.get": {
		"id", "name", "description", "createdAt",
	},

	// Initiative commands
	"initiative.list": {
		"id", "name", "description", "createdAt",
	},
	"initiative.get": {
		"id", "name", "description", "createdAt",
	},

	// Document commands
	"document.list": {
		"id", "title", "content", "createdAt",
	},
	"document.get": {
		"id", "title", "content", "createdAt",
	},

	// Template commands
	"template.list": {
		"id", "name", "description", "createdAt",
	},
	"template.get": {
		"id", "name", "description", "createdAt",
	},

	// Favorite commands (no list, only create/delete)

	// Reaction commands (no list, only create/delete)

	// Notification commands (no standard list/get pattern)

	// Organization command
	"organization": {
		"id", "name", "urlKey", "createdAt",
	},

	// Viewer command
	"viewer": {
		"id", "name", "email", "displayName", "active",
	},

	// Status command (no defaults needed - returns rate limit info)
}

// GetDefaults returns the default fields for a command, checking for
// config overrides before falling back to built-in defaults.
//
// Resolution order:
// 1. User config override (field_defaults in config file)
// 2. Built-in hardcoded defaults
// 3. nil (no defaults found)
//
// Note: Config parse errors are silently ignored - falls back to built-in defaults.
func GetDefaults(command string, configOverrides map[string]string) []string {
	// 1. Check user config override
	if configOverrides != nil {
		if override, ok := configOverrides[command]; ok {
			// Note: parseFields is in selector.go, we need to import it or duplicate logic
			// For now, parse inline to avoid circular dependency
			fields := parseConfigFields(override)
			if len(fields) > 0 {
				return fields
			}
		}
	}

	// 2. Fall back to built-in
	if defaults, ok := BuiltinDefaults[command]; ok {
		return defaults
	}

	// 3. No defaults found
	return nil
}

// GetDefaultsForMCP returns MCP-specific defaults, with fallback to
// regular defaults if MCP-specific config is not provided.
//
// Resolution order:
// 1. MCP-specific config override
// 2. Regular config override
// 3. Built-in hardcoded defaults
// 4. nil (no defaults found)
func GetDefaultsForMCP(command string, mcpOverrides, configOverrides map[string]string) []string {
	// 1. Check MCP-specific config
	if mcpOverrides != nil {
		if override, ok := mcpOverrides[command]; ok {
			fields := parseConfigFields(override)
			if len(fields) > 0 {
				return fields
			}
		}
	}

	// 2. Fall back to regular defaults
	return GetDefaults(command, configOverrides)
}

// parseConfigFields parses comma-separated field spec from config.
// Silently ignores empty fields (config is user-editable, be forgiving).
func parseConfigFields(fieldSpec string) []string {
	if fieldSpec == "" {
		return nil
	}

	var fields []string
	for field := range strings.SplitSeq(fieldSpec, ",") {
		field = strings.TrimSpace(field)
		if field != "" {
			fields = append(fields, field)
		}
	}
	return fields
}
