// Package organization provides organization commands for the Linear CLI.
package organization

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewOrganizationCommand creates the organization command.
func NewOrganizationCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "organization",
		Short:   "Get organization information",
		Aliases: []string{"org"},
		Long: `Get information about the current Linear organization/workspace.

Examples:
  linear organization
  linear org --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return run(cmd, client)
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func run(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	org, err := client.Organization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), org, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", org.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "URL:  %s\n", org.URLKey)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
