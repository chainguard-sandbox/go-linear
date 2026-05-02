package notification

import (
	"fmt"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
	"github.com/spf13/cobra"
)

// NewListCommand creates the notification list command.
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
	paginationFlags := &cli.PaginationFlags{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications (inbox)",
		Long: `List notifications from inbox. Returns 4 default fields per notification. Excludes archived by default.

Flags: --include-archived, --limit

Example: go-linear notification list
Example: go-linear notification list --include-archived --limit=100

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

	return formatter.FormatJSON(cmd.OutOrStdout(), notifications, true)
}
