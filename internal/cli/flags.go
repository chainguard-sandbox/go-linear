// Package cli provides shared types and utilities for CLI commands.
package cli

import (
	"github.com/spf13/cobra"
)

// FieldFlags holds the field selector flag for list/get commands.
type FieldFlags struct {
	Fields string // Field selector: defaults|none|defaults,extra|field1,field2
}

// Bind adds the fields flag to the command.
func (f *FieldFlags) Bind(cmd *cobra.Command, fieldsHelp string) {
	cmd.Flags().StringVar(&f.Fields, "fields", "", fieldsHelp)
}

// PaginationFlags holds common pagination flags.
type PaginationFlags struct {
	Limit int    // Number of items to return
	After string // Cursor for pagination
}

// Bind adds pagination flags to the command.
func (p *PaginationFlags) Bind(cmd *cobra.Command, defaultLimit int) {
	cmd.Flags().IntVar(&p.Limit, "limit", defaultLimit, "Number of items to return")
	cmd.Flags().StringVar(&p.After, "after", "", "Cursor for pagination")
}

// LimitPtr returns limit as *int64 for API calls.
func (p *PaginationFlags) LimitPtr() *int64 {
	limit := int64(p.Limit)
	return &limit
}

// AfterPtr returns after cursor as *string (nil if empty).
func (p *PaginationFlags) AfterPtr() *string {
	if p.After == "" {
		return nil
	}
	return &p.After
}

// ConfirmationFlags holds flags for destructive operations.
type ConfirmationFlags struct {
	Yes bool // Skip confirmation prompt
}

// Bind adds confirmation flags to the command.
func (c *ConfirmationFlags) Bind(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&c.Yes, "yes", false, "Skip confirmation prompt")
}
