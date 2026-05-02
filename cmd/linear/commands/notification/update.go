package notification

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/dateparser"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewUpdateCommand creates the notification update command.
func NewUpdateCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <notification-id>",
		Short: "Mark a notification as read or snooze it",
		Long: `Update notification. Modifies existing data.

Flags: --read (mark as read) | --snooze-until=tomorrow|3d (date formats: see issue_list)

Example: go-linear notification update <uuid> --snooze-until=tomorrow

Related: notification_archive, notification_subscribe`,
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

	return cmd
}

func runUpdate(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := cmd.Context()

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

	return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
}
