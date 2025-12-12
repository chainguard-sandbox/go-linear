// Package fieldfilter provides JSON field filtering for sparse output.
package fieldfilter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldSelector filters JSON objects to include only selected fields.
type FieldSelector struct {
	fields map[string]bool // "id", "title", "assignee.name"
}

// New creates a FieldSelector from comma-separated field specification.
// Examples: "id,title", "id,assignee.name,state.name"
// Returns nil FieldSelector if fieldSpec is empty (no filtering).
func New(fieldSpec string) (*FieldSelector, error) {
	if fieldSpec == "" {
		return nil, nil // nil means no filtering
	}

	fields := make(map[string]bool)
	for field := range strings.SplitSeq(fieldSpec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			return nil, fmt.Errorf("empty field name in spec")
		}
		fields[field] = true
	}

	return &FieldSelector{fields: fields}, nil
}

// Filter removes unselected fields from JSON data.
// If FieldSelector is nil, returns data unchanged.
func (fs *FieldSelector) Filter(data []byte) ([]byte, error) {
	if fs == nil {
		return data, nil // No filtering
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
	switch v := val.(type) {
	case map[string]any:
		return fs.filterObject(v)
	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = fs.filterValue(item)
		}
		return result
	default:
		return v
	}
}

// filterObject filters a single object to include only selected fields
func (fs *FieldSelector) filterObject(obj map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range obj {
		// Check if field is selected (direct match)
		if fs.fields[key] {
			result[key] = value
			continue
		}

		// Check for nested field selection (e.g., "assignee.name")
		hasNested := false
		prefix := key + "."
		for field := range fs.fields {
			if strings.HasPrefix(field, prefix) {
				hasNested = true
				break
			}
		}

		if hasNested {
			// Include parent object, filter recursively
			result[key] = fs.filterValue(value)
		}
	}

	return result
}
