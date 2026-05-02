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

// NewUnsnoozeAllCommand creates the notification unsnooze-all command.
func NewUnsnoozeAllCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsnooze-all",
		Short: "Unsnooze all notifications for an entity",
		Long: `Unsnooze all previously snoozed notifications for a specific entity.

Requires exactly one of: --issue, --project, --initiative, --notification

Example: go-linear notification unsnooze-all --issue=ENG-123

Related: notification_snooze-all, notification_archive-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUnsnoozeAll(cmd, client)
		},
	}

	addEntityFlags(cmd)
	return cmd
}

func runUnsnoozeAll(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input, err := buildEntityInput(cmd, ctx, res)
	if err != nil {
		return err
	}

	if err := client.NotificationUnsnoozeAll(ctx, input, time.Now()); err != nil {
		return fmt.Errorf("failed to unsnooze all notifications: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"action":  "unsnooze-all",
	}, true)
}
