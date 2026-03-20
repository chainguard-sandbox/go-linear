package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewAddLabelCommand creates the issue add-label command.
func NewAddLabelCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-label <issue-id> <label>",
		Short: "Add a label to an issue",
		Long: `Add label to issue. Safe operation.

Example: go-linear issue add-label ENG-123 bug

Related: issue_remove-label, label_list`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runAddLabel(cmd, client, args[0], args[1])
		},
	}

	return cmd
}

func runAddLabel(cmd *cobra.Command, client *linear.Client, issueID, labelName string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve issue ID
	resolvedIssueID, err := res.ResolveIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	// Resolve label name to UUID
	labelID, err := res.ResolveLabel(ctx, labelName)
	if err != nil {
		return fmt.Errorf("failed to resolve label: %w", err)
	}

	// Add label to issue
	issue, err := client.IssueAddLabel(ctx, resolvedIssueID, labelID)
	if err != nil {
		return fmt.Errorf("failed to add label: %w", err)
	}

	// Format output
	return formatter.FormatJSON(cmd.OutOrStdout(), issue, true)
}
