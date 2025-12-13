// Package organization provides organization commands for the Linear CLI.
package organization

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
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
  go-linear-cli organization
  go-linear-cli org --output=json`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,urlKey,createdAt) | none | defaults,extra")

	return cmd
}

func run(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	org, err := client.Organization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	fieldsSpec, _ := cmd.Flags().GetString("fields")

	switch output {
	case "json":
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
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", org.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "URL:  %s\n", org.URLKey)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
