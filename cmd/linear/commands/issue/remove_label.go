package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewRemoveLabelCommand creates the issue remove-label command.
func NewRemoveLabelCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-label <issue-id> <label>",
		Short: "Remove a label from an issue",
		Long: `Remove a label from an existing issue.

This operation is safe and can be reversed with 'issue add-label'.
Simpler than using 'issue update --remove-label' for single label operations.

Parameters:
  <issue-id>: Issue identifier (e.g., ENG-123) or UUID (required)
  <label>: Label name or UUID to remove (required)

Examples:
  # Remove 'bug' label from issue
  linear issue remove-label ENG-123 bug

  # Remove label using UUID
  linear issue remove-label <issue-uuid> <label-uuid>

  # Remove label with JSON output
  linear issue remove-label ENG-123 feature --output=json

TIP: Use 'linear issue get ENG-123 --output=json' to see current labels

Related Commands:
  - linear issue add-label - Add a label to an issue
  - linear issue update --remove-label - Alternative using update command
  - linear label list - List all available labels`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runRemoveLabel(cmd, client, args[0], args[1])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runRemoveLabel(cmd *cobra.Command, client *linear.Client, issueID, labelName string) error {
	ctx := context.Background()
	res := resolver.New(client)

	// Resolve label name to UUID
	labelID, err := res.ResolveLabel(ctx, labelName)
	if err != nil {
		return fmt.Errorf("failed to resolve label: %w", err)
	}

	// Remove label from issue
	issue, err := client.IssueRemoveLabel(ctx, issueID, labelID)
	if err != nil {
		return fmt.Errorf("failed to remove label: %w", err)
	}

	// Format output
	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), issue, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Removed label '%s' from issue %s\n", labelName, issueID)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
