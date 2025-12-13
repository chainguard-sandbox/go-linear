// Package formatter provides output formatting utilities for the Linear CLI.
package formatter

import (
	"encoding/json"
	"io"

	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
)

// FormatJSON writes data as JSON to the writer.
// If pretty is true, the JSON is indented for readability.
func FormatJSON(w io.Writer, data any, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// FormatJSONFiltered writes filtered JSON to writer.
// If fieldSelector is nil, behaves like FormatJSON (no filtering).
// Otherwise, filters JSON to include only selected fields.
func FormatJSONFiltered(w io.Writer, data any, pretty bool, fieldSelector *fieldfilter.FieldSelector) error {
	// Marshal to JSON first
	var jsonData []byte
	var err error

	if pretty {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}
	if err != nil {
		return err
	}

	// Apply field filtering if provided
	if fieldSelector != nil {
		jsonData, err = fieldSelector.Filter(jsonData)
		if err != nil {
			return err
		}
	}

	_, err = w.Write(jsonData)
	return err
}

// FormatJSONWithFields writes JSON with field filtering based on fieldSpec and defaults.
// This is a convenience wrapper that creates the FieldSelector from spec + defaults.
//
// Special field specs:
//   - "defaults" - use commandDefaults
//   - "none" or "" - no filtering (all fields)
//   - "defaults,field1,field2" - defaults + additional fields
//   - "field1,field2" - only specified fields
func FormatJSONWithFields(w io.Writer, data any, pretty bool, fieldSpec string, commandDefaults []string) error {
	// Create field selector with defaults support
	selector, err := fieldfilter.New(fieldSpec, commandDefaults)
	if err != nil {
		return err
	}

	return FormatJSONFiltered(w, data, pretty, selector)
}
