package formatter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
)

func TestFormatJSON(t *testing.T) {
	data := map[string]any{
		"id":    "123",
		"title": "Test",
	}

	tests := []struct {
		name   string
		pretty bool
	}{
		{"compact", false},
		{"pretty", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := FormatJSON(&buf, data, tt.pretty)
			if err != nil {
				t.Fatalf("FormatJSON() error = %v", err)
			}

			// Verify it's valid JSON
			var result map[string]any
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("FormatJSON() produced invalid JSON: %v", err)
			}

			if result["id"] != "123" || result["title"] != "Test" {
				t.Errorf("FormatJSON() = %v, want original data", result)
			}
		})
	}
}

func TestFormatJSONFiltered(t *testing.T) {
	data := map[string]any{
		"id":       "123",
		"title":    "Test",
		"priority": 1,
		"assignee": map[string]any{
			"name":  "Alice",
			"email": "alice@example.com",
		},
	}

	tests := []struct {
		name       string
		fieldSpec  string
		wantFields []string
	}{
		{
			name:       "nil selector",
			fieldSpec:  "",
			wantFields: []string{"id", "title", "priority", "assignee"},
		},
		{
			name:       "single field",
			fieldSpec:  "id",
			wantFields: []string{"id"},
		},
		{
			name:       "multiple fields",
			fieldSpec:  "id,title",
			wantFields: []string{"id", "title"},
		},
		{
			name:       "nested field",
			fieldSpec:  "id,assignee.name",
			wantFields: []string{"id", "assignee"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := fieldfilter.New(tt.fieldSpec, nil)
			if err != nil {
				t.Fatalf("fieldfilter.New() error = %v", err)
			}

			var buf bytes.Buffer
			err = FormatJSONFiltered(&buf, data, false, fs)
			if err != nil {
				t.Fatalf("FormatJSONFiltered() error = %v", err)
			}

			// Parse result
			var result map[string]any
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("FormatJSONFiltered() produced invalid JSON: %v", err)
			}

			// Check field count
			if len(result) != len(tt.wantFields) {
				t.Errorf("FormatJSONFiltered() field count = %d, want %d", len(result), len(tt.wantFields))
			}

			// Check expected fields present
			for _, field := range tt.wantFields {
				if _, ok := result[field]; !ok {
					t.Errorf("FormatJSONFiltered() missing field %s", field)
				}
			}
		})
	}
}
