package audit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewTypesCommand creates the audit types command.
func NewTypesCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "List available audit entry types",
		Long: `List all available audit entry types with descriptions.

Use this to discover valid values for the --type filter in audit list.

Example: go-linear audit types

Related: audit_list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runTypes(cmd, client)
		},
	}

	return cmd
}

func runTypes(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	types, err := client.AuditEntryTypes(ctx)
	if err != nil {
		return fmt.Errorf("failed to list audit entry types: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), types, true)
}
