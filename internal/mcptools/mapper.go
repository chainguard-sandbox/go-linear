// Package mcptools provides utilities for enhancing MCP tool definitions
package mcptools

import (
	"fmt"
	"strings"
)

// OutputSchemaMapping describes how a command maps to an output schema
type OutputSchemaMapping struct {
	CommandName        string
	PrimarySchema      string              // Main schema (e.g., "IssuesConnection")
	ConditionalSchemas []ConditionalSchema // Alternative schemas based on flags
	TableFormat        bool                // Whether table output is supported
	Description        string              // Human-readable description
}

// ConditionalSchema represents a schema that applies under certain conditions
type ConditionalSchema struct {
	Condition   string   // Human-readable condition (e.g., "when --group-by flags used")
	Flags       []string // Flags that trigger this schema
	SchemaName  string   // Schema to use when condition is met
	Description string   // Additional context
}

// InputSchema represents the structure of a tool's input schema from Ophis
type InputSchema struct {
	Type       string              `json:"type"`
	Required   []string            `json:"required"`
	Properties map[string]Property `json:"properties"`
}

// Property represents a property in the input schema
type Property struct {
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Default     any                 `json:"default,omitempty"`
	Properties  map[string]Property `json:"properties,omitempty"`
}

// PatternMapper maps commands to output schemas
type PatternMapper struct {
	specialCases map[string]string // Command-specific mappings
}

// NewPatternMapper creates a new pattern mapper
func NewPatternMapper() *PatternMapper {
	return &PatternMapper{
		specialCases: initSpecialCases(),
	}
}

// MapCommand maps a command to its output schema(s)
func (m *PatternMapper) MapCommand(toolName string, inputSchema InputSchema) OutputSchemaMapping {
	// Extract entity name (e.g., "linear_issue_list" → "Issue")
	entity := m.extractEntity(toolName)

	// Check for special cases first
	if schema, exists := m.specialCases[toolName]; exists {
		return OutputSchemaMapping{
			CommandName:   toolName,
			PrimarySchema: schema,
			TableFormat:   true,
			Description:   "Special case mapping",
		}
	}

	// Check for Phase 2 aggregation flags (group-by, time-based grouping)
	hasGroupBy := m.hasFlag(inputSchema, "group-by")
	hasTimeGrouping := m.hasFlag(inputSchema, "group-by-week") ||
		m.hasFlag(inputSchema, "group-by-month") ||
		m.hasFlag(inputSchema, "group-by-cycle")
	hasSorting := m.hasFlag(inputSchema, "sort-by")
	hasFiltering := m.hasFlag(inputSchema, "min-count") || m.hasFlag(inputSchema, "max-count")
	hasRegex := m.hasFlag(inputSchema, "regex")

	// Build conditional schemas for aggregation
	var conditionals []ConditionalSchema
	if hasGroupBy || hasTimeGrouping {
		flags := []string{}
		if hasGroupBy {
			flags = append(flags, "group-by")
		}
		if hasTimeGrouping {
			if m.hasFlag(inputSchema, "group-by-week") {
				flags = append(flags, "group-by-week")
			}
			if m.hasFlag(inputSchema, "group-by-month") {
				flags = append(flags, "group-by-month")
			}
			if m.hasFlag(inputSchema, "group-by-cycle") {
				flags = append(flags, "group-by-cycle")
			}
		}

		description := "Grouped and aggregated results"
		if hasSorting {
			description += ", sorted by specified field or count"
		}
		if hasFiltering {
			description += ", filtered by count range"
		}

		conditionals = append(conditionals, ConditionalSchema{
			Condition:   "any of --group-by, --group-by-week, --group-by-month, or --group-by-cycle is specified",
			Flags:       flags,
			SchemaName:  "GroupedResult",
			Description: description,
		})
	}

	// Detect command pattern
	var primarySchema string
	var description string

	switch {
	case strings.HasSuffix(toolName, "_list"):
		primarySchema = entity + "Connection"
		description = "Paginated list of " + pluralize(entity)
		if hasRegex {
			description += ", optionally filtered by regex pattern"
		}

	case strings.HasSuffix(toolName, "_search"):
		primarySchema = entity + "Connection"
		description = "Search results for " + pluralize(entity)
		if hasRegex {
			description += " with regex support"
		}

	case strings.HasSuffix(toolName, "_get"):
		primarySchema = entity
		description = "Single " + entity + " entity"

	case strings.HasSuffix(toolName, "_create"):
		primarySchema = entity + "Payload"
		description = "Result of creating a " + entity

	case strings.HasSuffix(toolName, "_update"):
		primarySchema = entity + "Payload"
		description = "Result of updating a " + entity

	case strings.HasSuffix(toolName, "_delete"):
		primarySchema = "ArchivePayload"
		description = "Result of deleting a " + entity

	case strings.HasSuffix(toolName, "_archive"):
		primarySchema = "ArchivePayload"
		description = "Result of archiving a " + entity

	case strings.HasSuffix(toolName, "_unarchive"):
		primarySchema = entity + "Payload"
		description = "Result of unarchiving a " + entity

	default:
		// Unknown pattern - use entity name
		primarySchema = entity
		description = entity + " entity"
	}

	return OutputSchemaMapping{
		CommandName:        toolName,
		PrimarySchema:      primarySchema,
		ConditionalSchemas: conditionals,
		TableFormat:        true,
		Description:        description,
	}
}

// pluralize returns the plural form of an entity name for descriptions
func pluralize(entity string) string {
	// Most entities just add 's'
	// Special cases for proper English pluralization
	switch entity {
	case "WorkflowState":
		return "workflow states"
	case "IssueLabel":
		return "issue labels"
	default:
		return entity + "s"
	}
}

// extractEntity extracts the entity name from a command name
// Examples: "linear_issue_list" → "Issue", "linear_team_members" → "Team"
func (m *PatternMapper) extractEntity(toolName string) string {
	// Remove "linear_" prefix
	name := strings.TrimPrefix(toolName, "linear_")

	// Split by underscore
	parts := strings.Split(name, "_")
	if len(parts) == 0 {
		return ""
	}

	// First part is usually the entity
	entityPart := parts[0]

	// Handle special naming cases
	switch entityPart {
	case "label":
		return "IssueLabel"
	case "state":
		return "WorkflowState"
	}

	// Capitalize first letter
	if entityPart != "" {
		return strings.ToUpper(entityPart[:1]) + entityPart[1:]
	}

	return entityPart
}

// hasFlag checks if a flag exists in the input schema
func (m *PatternMapper) hasFlag(inputSchema InputSchema, flagName string) bool {
	if inputSchema.Properties == nil {
		return false
	}

	flags, ok := inputSchema.Properties["flags"]
	if !ok || flags.Properties == nil {
		return false
	}

	_, exists := flags.Properties[flagName]
	return exists
}

// initSpecialCases returns command-specific schema mappings
func initSpecialCases() map[string]string {
	return map[string]string{
		// User commands that return issues
		"linear_user_completed": "IssueConnection",

		// Team commands that return specific types
		"linear_team_members":  "UserConnection",
		"linear_team_cycles":   "CycleConnection",
		"linear_team_projects": "ProjectConnection",
		"linear_team_stats":    "TeamStats", // Custom aggregation

		// Notification operations
		"linear_notification_subscribe":   "NotificationSubscriptionPayload",
		"linear_notification_unsubscribe": "ArchivePayload", // Returns success only (no payload)
		"linear_notification_archive":     "NotificationArchivePayload",

		// Project milestone operations (note: hyphens, not underscores)
		"linear_project_milestone-create": "ProjectMilestonePayload",
		"linear_project_milestone-update": "ProjectMilestonePayload",
		"linear_project_milestone-delete": "ArchivePayload",

		// Issue relation operations
		"linear_issue_relate":          "IssueRelationPayload",
		"linear_issue_unrelate":        "ArchivePayload",
		"linear_issue_update-relation": "IssueRelationPayload",

		// Attachment link operations (note: hyphens, not underscores)
		"linear_attachment_link-url":    "AttachmentPayload",
		"linear_attachment_link-github": "AttachmentPayload",
		"linear_attachment_link-slack":  "AttachmentPayload",

		// Issue label operations (note: hyphens, not underscores)
		"linear_issue_add-label":    "IssuePayload",
		"linear_issue_remove-label": "IssuePayload",

		// Favorite operations
		"linear_favorite_create": "FavoritePayload",
		"linear_favorite_delete": "ArchivePayload",

		// Reaction operations
		"linear_reaction_create": "ReactionPayload",
		"linear_reaction_delete": "ArchivePayload",

		// Organization view (single entity, not a list)
		"linear_organization": "Organization",

		// Viewer (current user)
		"linear_viewer": "User",
	}
}

// GetAllMappings returns mappings for all known commands
// This is useful for validation and testing
func (m *PatternMapper) GetAllMappings(commands []string, schemas map[string]InputSchema) []OutputSchemaMapping {
	mappings := make([]OutputSchemaMapping, 0, len(commands))

	for _, cmd := range commands {
		schema := InputSchema{}
		if s, exists := schemas[cmd]; exists {
			schema = s
		}

		mapping := m.MapCommand(cmd, schema)
		mappings = append(mappings, mapping)
	}

	return mappings
}

// ValidateMapping checks if a mapping's schemas exist in the schema definitions
func ValidateMapping(mapping OutputSchemaMapping, availableSchemas map[string]bool) []string {
	var missing []string

	// Check primary schema
	if !availableSchemas[mapping.PrimarySchema] {
		missing = append(missing, mapping.PrimarySchema)
	}

	// Check conditional schemas
	for _, conditional := range mapping.ConditionalSchemas {
		if !availableSchemas[conditional.SchemaName] {
			missing = append(missing, conditional.SchemaName)
		}
	}

	return missing
}

// FormatMappingReport generates a human-readable report of command mappings
func FormatMappingReport(mappings []OutputSchemaMapping) string {
	var report strings.Builder

	report.WriteString("Command Pattern Mapping Report\n")
	report.WriteString("================================\n\n")

	// Group by pattern
	patterns := make(map[string][]OutputSchemaMapping)
	for _, m := range mappings {
		pattern := "other"
		switch {
		case strings.HasSuffix(m.CommandName, "_list"):
			pattern = "list"
		case strings.HasSuffix(m.CommandName, "_get"):
			pattern = "get"
		case strings.HasSuffix(m.CommandName, "_create"):
			pattern = "create"
		case strings.HasSuffix(m.CommandName, "_update"):
			pattern = "update"
		case strings.HasSuffix(m.CommandName, "_delete"):
			pattern = "delete"
		case strings.HasSuffix(m.CommandName, "_search"):
			pattern = "search"
		}

		patterns[pattern] = append(patterns[pattern], m)
	}

	// Output grouped mappings
	for pattern, mappings := range patterns {
		fmt.Fprintf(&report, "## %s commands (%d)\n\n", strings.ToUpper(pattern), len(mappings))

		for _, m := range mappings {
			fmt.Fprintf(&report, "- %s → %s\n", m.CommandName, m.PrimarySchema)
			for _, cond := range m.ConditionalSchemas {
				fmt.Fprintf(&report, "  └─ %s → %s (flags: %s)\n",
					cond.Condition, cond.SchemaName, strings.Join(cond.Flags, ", "))
			}
		}
		report.WriteString("\n")
	}

	return report.String()
}
