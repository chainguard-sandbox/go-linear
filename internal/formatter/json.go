// Package formatter provides output formatting utilities for the Linear CLI.
package formatter

import (
	"encoding/json"
	"io"
)

// FormatJSON writes data as JSON to the writer.
// If pretty is true, the JSON is indented for readability.
func FormatJSON(w io.Writer, data interface{}, pretty bool) error {
	encoder := json.NewEncoder(w)
	if pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}
