package linear

import _ "embed"

// Schema contains the JSON Schema definitions for all go-linear types.
//
// This schema describes:
//   - Input types (IssueCreateInput, IssueUpdateInput, etc.)
//   - Output types (Issue, Team, WorkflowState, etc.)
//   - Method signatures and parameters
//   - Error types and handling
//
// The schema is useful for:
//   - AI agents understanding the API structure
//   - Generating client code in other languages
//   - Documentation and tooling
//   - MCP (Model Context Protocol) server integration
//
// Format: JSON Schema Draft 7
//
// Example usage:
//
//	schema := linear.GetSchema()
//	fmt.Println(schema)
//
// See also: schema.json in this package directory
//
//go:embed schema.json
var schema string

// GetSchema returns the JSON Schema for all go-linear types and methods.
//
// The returned string contains a complete JSON Schema document that describes:
//   - All input types with required fields and validation rules
//   - All output types with field descriptions and nullability
//   - Method signatures with parameters and return types
//   - Common workflows and error scenarios
//
// This is particularly useful for AI agents and tools that need to understand
// the Linear API structure programmatically.
//
// Example:
//
//	import "encoding/json"
//	import "github.com/chainguard-sandbox/go-linear/pkg/linear"
//
//	schema := linear.GetSchema()
//	var parsed map[string]interface{}
//	json.Unmarshal([]byte(schema), &parsed)
func GetSchema() string {
	return schema
}
