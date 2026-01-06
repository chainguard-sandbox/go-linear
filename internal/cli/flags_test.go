package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputFlags_Bind(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := &OutputFlags{}

	flags.Bind(cmd, "Field selector help text")

	// Check output flag exists with correct default
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
	if outputFlag.DefValue != "table" {
		t.Errorf("output default = %q, want %q", outputFlag.DefValue, "table")
	}
	if outputFlag.Shorthand != "o" {
		t.Errorf("output shorthand = %q, want %q", outputFlag.Shorthand, "o")
	}

	// Check fields flag exists
	fieldsFlag := cmd.Flags().Lookup("fields")
	if fieldsFlag == nil {
		t.Fatal("fields flag not found")
	}
	if fieldsFlag.Usage != "Field selector help text" {
		t.Errorf("fields usage = %q, want custom help text", fieldsFlag.Usage)
	}
}

func TestOutputFlags_Validate(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		wantErr bool
	}{
		{"json is valid", "json", false},
		{"table is valid", "table", false},
		{"empty is invalid", "", true},
		{"yaml is invalid", "yaml", true},
		{"csv is invalid", "csv", true},
		{"JSON uppercase is invalid", "JSON", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &OutputFlags{Output: tt.output}
			err := f.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPaginationFlags_Bind(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := &PaginationFlags{}

	flags.Bind(cmd, 100)

	// Check limit flag
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("limit flag not found")
	}
	if limitFlag.DefValue != "100" {
		t.Errorf("limit default = %q, want %q", limitFlag.DefValue, "100")
	}
	if limitFlag.Shorthand != "" {
		t.Errorf("limit shorthand = %q, want %q (no shorthand to avoid conflicts)", limitFlag.Shorthand, "")
	}

	// Check after flag
	afterFlag := cmd.Flags().Lookup("after")
	if afterFlag == nil {
		t.Fatal("after flag not found")
	}
}

func TestPaginationFlags_LimitPtr(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int64
	}{
		{"zero", 0, 0},
		{"positive", 50, 50},
		{"large", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaginationFlags{Limit: tt.limit}
			got := p.LimitPtr()
			if got == nil {
				t.Fatal("LimitPtr() returned nil")
			}
			if *got != tt.want {
				t.Errorf("LimitPtr() = %d, want %d", *got, tt.want)
			}
		})
	}
}

func TestPaginationFlags_AfterPtr(t *testing.T) {
	tests := []struct {
		name    string
		after   string
		wantNil bool
	}{
		{"empty returns nil", "", true},
		{"cursor returns pointer", "cursor123", false},
		{"whitespace returns pointer", " ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaginationFlags{After: tt.after}
			got := p.AfterPtr()
			if tt.wantNil {
				if got != nil {
					t.Errorf("AfterPtr() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Fatal("AfterPtr() returned nil, want pointer")
				}
				if *got != tt.after {
					t.Errorf("AfterPtr() = %q, want %q", *got, tt.after)
				}
			}
		})
	}
}

func TestConfirmationFlags_Bind(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	flags := &ConfirmationFlags{}

	flags.Bind(cmd)

	// Check yes flag
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("yes flag not found")
	}
	if yesFlag.DefValue != "false" {
		t.Errorf("yes default = %q, want %q", yesFlag.DefValue, "false")
	}
}

// TestFlags_Integration tests that flags work together in a command execution.
func TestFlags_Integration(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "test",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}

	output := &OutputFlags{}
	pagination := &PaginationFlags{}
	confirm := &ConfirmationFlags{}

	output.Bind(cmd, "Select fields")
	pagination.Bind(cmd, 50)
	confirm.Bind(cmd)

	// Simulate parsing flags
	cmd.SetArgs([]string{"--output=json", "--limit=25", "--after=abc123", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// Verify flag values were captured
	if output.Output != "json" {
		t.Errorf("output = %q, want %q", output.Output, "json")
	}
	if pagination.Limit != 25 {
		t.Errorf("limit = %d, want %d", pagination.Limit, 25)
	}
	if pagination.After != "abc123" {
		t.Errorf("after = %q, want %q", pagination.After, "abc123")
	}
	if !confirm.Yes {
		t.Error("yes = false, want true")
	}

	// Verify helper methods
	if err := output.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
	if *pagination.LimitPtr() != 25 {
		t.Errorf("LimitPtr() = %d, want %d", *pagination.LimitPtr(), 25)
	}
	if *pagination.AfterPtr() != "abc123" {
		t.Errorf("AfterPtr() = %q, want %q", *pagination.AfterPtr(), "abc123")
	}
}
