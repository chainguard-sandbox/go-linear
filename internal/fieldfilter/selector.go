// Package fieldfilter provides JSON field filtering for sparse output.
package fieldfilter

import (
	"bytes"
	"encoding/json"
	"strings"
)

// FieldSelector filters JSON objects to include only selected fields.
type FieldSelector struct {
	fields         map[string]bool // "id", "title", "assignee.name"
	preserveFields map[string]bool // Always preserve these (e.g., "nodes", "pageInfo")
	nullable       map[string]bool // Selected fields emitted as null when absent from a list item
	passAll        bool            // Keep every field; only pad nullable fields in top-level nodes
}

// WithNullable marks fields that should be emitted as JSON null when they are
// absent from a list item. Generated types omit nil pointers (omitempty), so
// without this an issue with no SLA would drop slaBreachesAt entirely rather
// than report it as null. Pagination wrappers like pageInfo are never padded.
//
// On a selector built from a field list, only selected nullable fields are
// padded, in items inside arrays. On a nil selector ("none", or no defaults),
// every field is kept and all nullable fields are padded in the items of the
// top-level "nodes" array, so "none" reports unset fields the same way an
// explicit field list does.
func (fs *FieldSelector) WithNullable(fields ...string) *FieldSelector {
	if len(fields) == 0 {
		return fs
	}
	if fs == nil {
		fs = &FieldSelector{passAll: true}
	}
	if fs.nullable == nil {
		fs.nullable = make(map[string]bool, len(fields))
	}
	for _, f := range fields {
		fs.nullable[f] = true
	}
	return fs
}

// New creates a FieldSelector from comma-separated field specification.
//
// Special values:
//   - "defaults" uses commandDefaults
//   - "none" or "" returns nil (no filtering)
//   - "defaults,field1,field2" merges defaults with additional fields
//
// Examples:
//   - "id,title" - only these fields
//   - "defaults" - use command defaults
//   - "defaults,description" - defaults + description
//   - "none" or "" - no filtering (all fields)
//
// Returns nil FieldSelector if no filtering should be applied.
func New(fieldSpec string, commandDefaults []string) (*FieldSelector, error) {
	// Handle "none" keyword - show all fields
	if fieldSpec == "none" {
		return nil, nil
	}

	// Empty spec with defaults - apply defaults
	if fieldSpec == "" {
		if len(commandDefaults) == 0 {
			return nil, nil // No defaults, no filtering
		}
		return newSelector(commandDefaults), nil
	}

	// Check if spec starts with "defaults"
	if additional, ok := strings.CutPrefix(fieldSpec, "defaults"); ok {
		// Parse: "defaults,body,estimate" → use defaults + [body, estimate]
		additional = strings.TrimPrefix(additional, ",")
		additional = strings.TrimSpace(additional)

		// Merge defaults with additional fields
		var fields []string
		fields = append(fields, commandDefaults...)

		if additional != "" {
			additionalFields, err := parseFields(additional)
			if err != nil {
				return nil, err
			}
			fields = append(fields, additionalFields...)
		}

		// Deduplicate
		fields = deduplicate(fields)

		return newSelector(fields), nil
	}

	// Explicit field list (no defaults)
	fields, err := parseFields(fieldSpec)
	if err != nil {
		return nil, err
	}
	return newSelector(fields), nil
}

// parseFields parses a comma-separated field specification into a slice.
// Returns error if any field is empty (e.g., "id,,title").
func parseFields(fieldSpec string) ([]string, error) {
	if fieldSpec == "" {
		return nil, nil
	}

	// Estimate capacity (count commas + 1)
	capacity := strings.Count(fieldSpec, ",") + 1
	fields := make([]string, 0, capacity)

	for field := range strings.SplitSeq(fieldSpec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			// Empty field detected (e.g., "id,,title")
			return nil, errEmptyField
		}
		fields = append(fields, field)
	}
	return fields, nil
}

var errEmptyField = func() error {
	// Using a closure to avoid importing "errors" or "fmt" just for this
	return &fieldError{msg: "empty field name in spec"}
}()

type fieldError struct {
	msg string
}

func (e *fieldError) Error() string {
	return e.msg
}

// deduplicate removes duplicate fields while preserving order.
func deduplicate(fields []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(fields))

	for _, field := range fields {
		if !seen[field] {
			seen[field] = true
			result = append(result, field)
		}
	}

	return result
}

// newSelector creates a FieldSelector from a field list.
func newSelector(fields []string) *FieldSelector {
	if len(fields) == 0 {
		return nil
	}

	fieldMap := make(map[string]bool, len(fields))
	for _, field := range fields {
		fieldMap[field] = true
	}

	return &FieldSelector{fields: fieldMap}
}

// NewForList creates a FieldSelector for list commands that always preserves
// the pagination wrapper structure (nodes, pageInfo).
// The field filtering applies to items inside the nodes array.
func NewForList(fieldSpec string, commandDefaults []string) (*FieldSelector, error) {
	fs, err := New(fieldSpec, commandDefaults)
	if err != nil {
		return nil, err
	}

	// No filtering requested
	if fs == nil {
		return nil, nil
	}

	// Always preserve pagination wrapper fields
	fs.preserveFields = map[string]bool{
		"nodes":      true,
		"pageInfo":   true,
		"totalCount": true,
	}

	return fs, nil
}

// Filter removes unselected fields from JSON data.
// If FieldSelector is nil, returns data unchanged.
func (fs *FieldSelector) Filter(data []byte) ([]byte, error) {
	if fs == nil {
		return data, nil // No filtering
	}
	if fs.passAll {
		return fs.padNodes(data)
	}

	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	filtered := fs.filterValue(obj)
	return json.Marshal(filtered)
}

// filterValue recursively filters map[string]any structures
func (fs *FieldSelector) filterValue(val any) any {
	if fs == nil {
		return val // No filtering if selector is nil
	}
	switch v := val.(type) {
	case map[string]any:
		return fs.filterObject(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			if obj, ok := item.(map[string]any); ok {
				result[i] = fs.padNullable(fs.filterObject(obj))
				continue
			}
			result[i] = fs.filterValue(item)
		}
		return result
	default:
		return v
	}
}

// padNodes keeps every field and adds a null entry for each nullable field
// missing from an item of the top-level "nodes" array. Nested arrays (such as
// labels.nodes) are left alone. Indented input stays indented.
func (fs *FieldSelector) padNodes(data []byte) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var obj any
	if err := dec.Decode(&obj); err != nil {
		return nil, err
	}
	wrapper, ok := obj.(map[string]any)
	if !ok {
		return data, nil
	}
	nodes, ok := wrapper["nodes"].([]any)
	if !ok {
		return data, nil
	}
	for _, item := range nodes {
		node, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for field := range fs.nullable {
			if _, present := node[field]; !present {
				node[field] = nil
			}
		}
	}
	if bytes.HasPrefix(bytes.TrimSpace(data), []byte("{\n")) {
		return json.MarshalIndent(wrapper, "", "  ")
	}
	return json.Marshal(wrapper)
}

// padNullable adds a null entry for each selected nullable field missing from obj.
func (fs *FieldSelector) padNullable(obj map[string]any) map[string]any {
	for field := range fs.nullable {
		if !fs.fields[field] {
			continue
		}
		if _, ok := obj[field]; !ok {
			obj[field] = nil
		}
	}
	return obj
}

// filterObject filters a single object to include only selected fields
func (fs *FieldSelector) filterObject(obj map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range obj {
		// Always preserve list wrapper fields. Field selection applies to the
		// items in "nodes"; pageInfo and totalCount pass through untouched so
		// pagination cursors survive --fields.
		if fs.preserveFields != nil && fs.preserveFields[key] {
			if key == "nodes" {
				result[key] = fs.filterValue(value)
			} else {
				result[key] = value
			}
			continue
		}

		// Check if field is selected (direct match)
		if fs.fields[key] {
			result[key] = value
			continue
		}

		// Check for nested field selection (e.g., "assignee.name")
		hasNested := false
		nestedFields := make(map[string]bool)
		prefix := key + "."
		for field := range fs.fields {
			if strings.HasPrefix(field, prefix) {
				hasNested = true
				// Extract the nested part (e.g., "assignee.name" → "name")
				nestedField := strings.TrimPrefix(field, prefix)
				nestedFields[nestedField] = true
			}
		}

		if hasNested {
			// Create a new field selector for the nested object
			nestedSelector := &FieldSelector{fields: nestedFields}
			result[key] = nestedSelector.filterValue(value)
		}
	}

	return result
}
