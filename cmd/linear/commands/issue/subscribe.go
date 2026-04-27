package issue

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewSubscribeCommand creates the issue subscribe command.
func NewSubscribeCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe <id>",
		Short: "Subscribe to issue notifications",
		Long: `Subscribe to an issue to receive notifications on updates.

Subscribes the current user by default. Use --user to subscribe another user.

Example: go-linear issue subscribe ENG-123

Related: issue_unsubscribe, notification_list`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSubscribe(cmd, client, args[0])
		},
	}

	cmd.Flags().String("user", "", "User to subscribe (name, email, or ID; default: current user)")
	return cmd
}

func runSubscribe(cmd *cobra.Command, client *linear.Client, issueID string) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	resolvedIssueID, err := res.ResolveIssue(ctx, issueID)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	var userIDPtr *string
	userFlag, _ := cmd.Flags().GetString("user")
	if userFlag != "" {
		userID, err := res.ResolveUser(ctx, userFlag)
		if err != nil {
			return fmt.Errorf("failed to resolve user: %w", err)
		}
		userIDPtr = &userID
	}

	if err := client.IssueSubscribe(ctx, resolvedIssueID, userIDPtr); err != nil {
		return fmt.Errorf("failed to subscribe to issue: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"issueId": issueID,
		"action":  "subscribed",
	}, true)
}
