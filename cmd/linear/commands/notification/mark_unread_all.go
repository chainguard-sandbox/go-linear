package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewMarkUnreadAllCommand creates the notification mark-unread-all command.
func NewMarkUnreadAllCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-unread-all",
		Short: "Mark all notifications as unread for an entity",
		Long: `Mark all notifications as unread for a specific entity (issue, project, initiative).

Requires exactly one of: --issue, --project, --initiative, --notification

Example: go-linear notification mark-unread-all --issue=ENG-123

Related: notification_mark-read-all, notification_archive-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMarkUnreadAll(cmd, client)
		},
	}

	addEntityFlags(cmd)
	cmd.Flags().Bool("yes", false, "Confirm bulk mark-unread without prompt")
	return cmd
}

func runMarkUnreadAll(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	yes, _ := cmd.Flags().GetBool("yes")
	if !yes {
		return fmt.Errorf("mark-unread-all marks all notifications as unread for the entity; pass --yes to confirm")
	}

	input, err := buildEntityInput(cmd, ctx, res)
	if err != nil {
		return err
	}

	if err := client.NotificationMarkUnreadAll(ctx, input); err != nil {
		return fmt.Errorf("failed to mark all notifications as unread: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"action":  "mark-unread-all",
	}, true)
}
