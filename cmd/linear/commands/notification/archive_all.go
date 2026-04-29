package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// NewArchiveAllCommand creates the notification archive-all command.
func NewArchiveAllCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive-all",
		Short: "Archive all notifications for an entity",
		Long: `Archive all notifications related to a specific entity (issue, project, initiative).

Requires exactly one of: --issue, --project, --initiative, --notification

Example: go-linear notification archive-all --issue=ENG-123

Related: notification_archive, notification_mark-read-all`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runArchiveAll(cmd, client)
		},
	}

	addEntityFlags(cmd)
	return cmd
}

func runArchiveAll(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()
	res := resolver.New(client)

	input, err := buildEntityInput(cmd, ctx, res)
	if err != nil {
		return err
	}

	if err := client.NotificationArchiveAll(ctx, input); err != nil {
		return fmt.Errorf("failed to archive all notifications: %w", err)
	}

	return formatter.FormatJSON(cmd.OutOrStdout(), map[string]any{
		"success": true,
		"action":  "archive-all",
	}, true)
}

