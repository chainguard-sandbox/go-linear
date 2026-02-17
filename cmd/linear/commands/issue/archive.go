package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewArchiveCommand creates the issue archive command.
func NewArchiveCommand(clientFactory cli.ClientFactory) *cobra.Command {
	var trash bool
	cmd := &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive an issue",
		Long: `Archive issue. Hides from default views. Can be unarchived.

Use --trash to move to trash instead (30-day auto-delete).

Example: go-linear issue archive ENG-123

Related: issue_unarchive, issue_delete, issue_get`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runArchive(cmd, client, args[0], trash)
		},
	}

	cmd.Flags().BoolVar(&trash, "trash", false, "Move to trash instead of archiving (30-day auto-delete)")
	return cmd
}

func runArchive(cmd *cobra.Command, client *linear.Client, issueID string, trash bool) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	// Resolve issue ID
	resolvedIssueID, err := res.ResolveIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	// Archive issue
	var trashPtr *bool
	if trash {
		trashPtr = &trash
	}
	err = client.IssueArchive(ctx, resolvedIssueID, trashPtr)
	if err != nil {
		return fmt.Errorf("failed to archive issue: %w", err)
	}

	// Format output
	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"issueId": issueID,
		"trashed": trash,
	}, true)
}
