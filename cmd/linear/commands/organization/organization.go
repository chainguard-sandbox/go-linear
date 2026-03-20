// Package organization provides organization commands for the Linear CLI.
package organization

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/config"
	"github.com/chainguard-sandbox/go-linear/v2/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewOrganizationCommand creates the organization command.
func NewOrganizationCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "organization",
		Short:   "Get organization information",
		Aliases: []string{"org"},
		Long: `Get information about the current Linear organization/workspace.

Examples:
  go-linear organization
  go-linear org`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return run(cmd, client)
		},
	}

	cmd.Flags().String("fields", "", "defaults (id,name,urlKey,createdAt) | none | defaults,extra")

	return cmd
}

func run(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	org, err := client.Organization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	fieldsSpec, _ := cmd.Flags().GetString("fields")

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("organization", configOverrides)
	fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), org, true, fieldSelector)
}
