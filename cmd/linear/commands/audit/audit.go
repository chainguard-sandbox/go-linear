// Package audit provides audit log commands for the Linear CLI.
package audit

import (
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
)

// NewAuditCommand creates the audit command group.
func NewAuditCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "View audit logs",
		Long:  "Commands for viewing audit log entries and available audit entry types.",
	}

	cmd.AddCommand(NewListCommand(clientFactory))
	cmd.AddCommand(NewTypesCommand(clientFactory))

	return cmd
}
