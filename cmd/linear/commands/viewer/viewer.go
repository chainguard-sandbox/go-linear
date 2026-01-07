// Package viewer provides commands for getting current user information.
package viewer

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/config"
	"github.com/chainguard-sandbox/go-linear/internal/fieldfilter"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// cli.ClientFactory is a function that creates a Linear client.

// NewViewerCommand creates the viewer command.
func NewViewerCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "viewer",
		Short: "Get current authenticated user information",
		Long: `Get information about the currently authenticated user.

Examples:
  go-linear viewer
  go-linear viewer --output=json`,
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
	cmd.Flags().String("fields", "", "defaults (id,name,email,displayName,active) | none | defaults,extra")

	return cmd
}

func run(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	viewer, err := client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get viewer: %w", err)
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
		defaults := fieldfilter.GetDefaults("viewer", configOverrides)
		fieldSelector, err := fieldfilter.New(fieldsSpec, defaults)
		if err != nil {
			return fmt.Errorf("invalid --fields: %w", err)
		}
		return formatter.FormatJSONFiltered(cmd.OutOrStdout(), viewer, true, fieldSelector)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "Name:   %s\n", viewer.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "Email:  %s\n", viewer.Email)
		fmt.Fprintf(cmd.OutOrStdout(), "ID:     %s\n", viewer.ID)
		if viewer.Admin {
			fmt.Fprintf(cmd.OutOrStdout(), "Admin:  Yes\n")
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
