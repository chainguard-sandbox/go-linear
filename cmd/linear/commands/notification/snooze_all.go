package notification

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewSnoozeAllCommand creates the notification snooze-all command.
func NewSnoozeAllCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snooze-all",
		Short: "Snooze all notifications for an entity",
		Long: `Snooze all notifications for a specific entity until a given time.

Requires exactly one of: --issue, --project, --initiative, --notification
Requires --until to specify when notifications should reappear.

Example: go-linear notification snooze-all --issue=ENG-123 --until=tomorrow
Example: go-linear notification snooze-all --issue=ENG-123 --until=3d

Related: notification_unsnooze-all, notification_archive-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSnoozeAll(cmd, client)
		},
	}

	addEntityFlags(cmd)
	cmd.Flags().String("until", "", "Snooze until (ISO8601, 'tomorrow', '3d', '1h')")
	_ = cmd.MarkFlagRequired("until")
	return cmd
}

func runSnoozeAll(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	input, err := buildEntityInput(cmd)
	if err != nil {
		return err
	}

	untilStr, _ := cmd.Flags().GetString("until")
	parser := dateparser.New()
	until, err := parser.Parse(untilStr)
	if err != nil {
		return fmt.Errorf("invalid --until value: %w", err)
	}

	if until.Before(time.Now()) {
		return fmt.Errorf("--until must be in the future")
	}

	if err := client.NotificationSnoozeAll(ctx, input, until); err != nil {
		return fmt.Errorf("failed to snooze all notifications: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success":        true,
		"action":         "snooze-all",
		"snoozedUntilAt": until.Format(time.RFC3339),
	}, true)
}
