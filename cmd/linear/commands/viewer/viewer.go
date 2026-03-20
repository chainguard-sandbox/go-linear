// Package viewer provides commands for getting current user information.
package viewer

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

// NewViewerCommand creates the viewer command.
func NewViewerCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "viewer",
		Short: "Get current authenticated user information",
		Long: `Get information about the currently authenticated user.

Examples:
  go-linear viewer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return run(cmd, client)
		},
	}

	cmd.Flags().String("fields", "", "defaults (id,name,email,displayName,active) | none | defaults,extra")

	return cmd
}

func run(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	viewer, err := client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get viewer: %w", err)
	}

	fieldsSpec, _ := cmd.Flags().GetString("fields")

	cfg, _ := config.Load()
	var configOverrides map[string]string
	if cfg != nil {
		configOverrides = cfg.FieldDefaults
	}
	defaults := fieldfilter.GetDefaults("viewer", configOverrides)
	fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
	if err != nil {
		return fmt.Errorf("invalid --fields: %w", err)
	}
	return formatter.FormatJSONFiltered(cmd.OutOrStdout(), viewer, true, fieldSelector)
}
