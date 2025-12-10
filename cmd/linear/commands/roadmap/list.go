package roadmap

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the roadmap list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all roadmaps",
		Long: `List all roadmaps in the Linear workspace.

Examples:
  linear roadmap list
  linear roadmap list --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 50, "Number of roadmaps to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	roadmaps, err := client.Roadmaps(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list roadmaps: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), roadmaps, true)
	case "table":
		if len(roadmaps.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No roadmaps found")
			return nil
		}
		for _, rm := range roadmaps.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", rm.Name)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
