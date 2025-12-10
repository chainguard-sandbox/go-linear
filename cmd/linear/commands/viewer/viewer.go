// Package viewer provides commands for getting current user information.
package viewer

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// ClientFactory is a function that creates a Linear client.
type ClientFactory func() (*linear.Client, error)

// NewViewerCommand creates the viewer command.
func NewViewerCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "viewer",
		Short: "Get current authenticated user information",
		Long: `Get information about the currently authenticated user.

Examples:
  linear viewer
  linear viewer --output=json`,
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

	viewer, err := client.Viewer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get viewer: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), viewer, true)
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
