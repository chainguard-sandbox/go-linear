package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewGetCommand creates the notification get command.
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <notification-id>",
		Short: "Get a single notification by ID",
		Long: `Get notification by UUID. Returns type, dates, read/snoozed status.

Example: go-linear notification get <uuid> --output=json

Related: notification_list, notification_archive`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runGet(cmd, client, args[0])
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := cmd.Context()

	notification, err := client.Notification(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to get notification: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), notification, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "ID:      %s\n", notification.ID)
		fmt.Fprintf(cmd.OutOrStdout(), "Type:    %s\n", notification.Type)
		fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", notification.CreatedAt)

		if notification.ReadAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Read:    %s\n", notification.ReadAt.Format("2006-01-02 15:04"))
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Read:    unread\n")
		}

		if notification.SnoozedUntilAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Snoozed: until %s\n", notification.SnoozedUntilAt.Format("2006-01-02 15:04"))
		}

		if notification.ArchivedAt != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Archived: %s\n", notification.ArchivedAt.Format("2006-01-02"))
		}

		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
