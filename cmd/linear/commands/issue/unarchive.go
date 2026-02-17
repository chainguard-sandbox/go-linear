package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUnarchiveCommand creates the issue unarchive command.
func NewUnarchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive <id>",
		Short: "Unarchive an issue",
		Long: `Restore archived or trashed issue. Safe operation.

Example: go-linear issue unarchive ENG-123

Related: issue_archive, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnarchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runUnarchive(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve issue ID
	resolvedIssueID, err := res.ResolveIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	// Unarchive issue
	err = client.IssueUnarchive(ctx, resolvedIssueID)
	if err != nil {
		return fmt.Errorf("failed to unarchive issue: %w", err)
	}

	// Format output
	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"issueId": issueID,
	}, true)
}
