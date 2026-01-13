package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewListCommand creates the notification list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications (inbox)",
		Long: `List notifications from inbox. Returns 4 default fields per notification. Excludes archived by default.

Flags: --include-archived, --limit

Example: go-linear notification list --output=json
Example: go-linear notification list --include-archived --limit=100 --output=json

Related: notification_get, notification_archive, notification_update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runList(cmd, client, paginationFlags)
		},
	}

	cmd.Flags().Bool("include-archived", false, "Include archived notifications")
	paginationFlags.Bind(cmd, 50)
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, paginationFlags *cli.PaginationFlags) error {
	ctx := cmd.Context()

	first := paginationFlags.LimitPtr()
	includeArchived, _ := cmd.Flags().GetBool("include-archived")

	var filter *intgraphql.NotificationFilter
	if !includeArchived {
		// Default: exclude archived
		filter = &intgraphql.NotificationFilter{}
	}

	notifications, err := client.Notifications(ctx, first, nil, filter)
	if err != nil {
		return fmt.Errorf("failed to list notifications: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), notifications, true)
	case "table":
		if len(notifications.Nodes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No notifications")
			return nil
		}
		for _, notif := range notifications.Nodes {
			status := ""
			if notif.ReadAt != nil {
				status = " [read]"
			} else if notif.SnoozedUntilAt != nil {
				status = " [snoozed]"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s%s\n", notif.Type, notif.ID, status)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
