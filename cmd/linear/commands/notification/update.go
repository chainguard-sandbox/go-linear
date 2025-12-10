package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewUpdateCommand creates the notification update command.
func NewUpdateCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <notification-id>",
		Short: "Mark a notification as read or snooze it",
		Long: `Update notification status (mark as read or snooze).

⚠️ CAUTION: This modifies notification state. Changes are immediate.

Use this to manage notification inbox programmatically.
Mark notifications as read after processing or snooze for later review.

Parameters:
  <notification-id>: Notification UUID (required)
  --read: Mark as read (sets readAt to current time)
  --snooze-until: Snooze until date/time (ISO8601, relative like "tomorrow", "3d")

Examples:
  # Mark notification as read
  linear notification update <notif-uuid> --read

  # Snooze until tomorrow
  linear notification update <notif-uuid> --snooze-until=tomorrow

  # Snooze for 3 days
  linear notification update <notif-uuid> --snooze-until=3d --output=json

TIP: Use snoozed notifications to defer non-urgent items

Related Commands:
  - linear notification archive - Archive notification (alternative)
  - linear notification subscribe - Subscribe to issue/project updates`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runUpdate(cmd, client, args[0])
		},
	}

	cmd.Flags().Bool("read", false, "Mark notification as read")
	cmd.Flags().String("snooze-until", "", "Snooze until date/time (ISO8601 or relative)")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := context.Background()

	read, _ := cmd.Flags().GetBool("read")
	snoozeUntilStr, _ := cmd.Flags().GetString("snooze-until")

	input := intgraphql.NotificationUpdateInput{}

	if read {
		now := time.Now()
		input.ReadAt = &now
	}

	if snoozeUntilStr != "" {
		parser := dateparser.New()
		snoozeUntil, err := parser.Parse(snoozeUntilStr)
		if err != nil {
			return fmt.Errorf("invalid snooze-until date: %w", err)
		}
		input.SnoozedUntilAt = &snoozeUntil
	}

	result, err := client.NotificationUpdate(ctx, notificationID, input)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Updated notification\n")
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
