package notification

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewArchiveCommand creates the notification archive command.
func NewArchiveCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <notification-id>",
		Short: "Archive a notification",
		Long: `Archive notification. Hides from inbox. Can be restored in UI.

Example: go-linear notification archive <uuid>

Related: notification_update`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runArchive(cmd, client, args[0])
		},
	}

	return cmd
}

func runArchive(cmd *cobra.Command, client *linear.Client, notificationID string) error {
	ctx := context.Background()

	err := client.NotificationArchive(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("failed to archive notification: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ Archived notification\n")
	return nil
}
