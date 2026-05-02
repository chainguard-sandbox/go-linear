package notification

import (
	"fmt"
	"time"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
	"github.com/spf13/cobra"
)

// NewMarkReadAllCommand creates the notification mark-read-all command.
func NewMarkReadAllCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-read-all",
		Short: "Mark all notifications as read for an entity",
		Long: `Mark all notifications as read for a specific entity (issue, project, initiative).

Requires exactly one of: --issue, --project, --initiative, --notification

Example: go-linear notification mark-read-all --issue=ENG-123

Related: notification_mark-unread-all, notification_archive-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runMarkReadAll(cmd, client)
		},
	}

	addEntityFlags(cmd)
	return cmd
}

func runMarkReadAll(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input, err := buildEntityInput(cmd, ctx, res)
	if err != nil {
		return err
	}

	if err := client.NotificationMarkReadAll(ctx, input, time.Now()); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"action":  "mark-read-all",
	}, true)
}
