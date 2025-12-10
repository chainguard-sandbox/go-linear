package label

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the label list command.
func NewListCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all issue labels",
		Long: `List all issue labels in the Linear workspace.

Examples:
  linear label list
  linear label list --output=json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client)
		},
	}

	cmd.Flags().IntP("limit", "l", 250, "Number of labels to return")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	limit, _ := cmd.Flags().GetInt("limit")
	first := int64(limit)

	labels, err := client.IssueLabels(ctx, &first, nil)
	if err != nil {
		return fmt.Errorf("failed to list labels: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), labels, true)
	case "table":
		if len(labels.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No labels found")
			return nil
		}
		for _, label := range labels.Nodes {
			fmt.Fprintf(cmd.OutOrStdout(), "%s (%s)\n", label.Name, label.Color)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
