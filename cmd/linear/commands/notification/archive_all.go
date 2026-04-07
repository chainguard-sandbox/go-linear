package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
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

	input, err := buildEntityInput(cmd)
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

// addEntityFlags adds the common entity filter flags for bulk notification operations.
func addEntityFlags(cmd *cobra.Command) {
	cmd.Flags().String("issue", "", "Issue ID (e.g., ENG-123 or UUID)")
	cmd.Flags().String("project", "", "Project ID (UUID)")
	cmd.Flags().String("initiative", "", "Initiative ID (UUID)")
	cmd.Flags().String("notification", "", "Notification ID (UUID)")
}

// buildEntityInput constructs a NotificationEntityInput from flags.
func buildEntityInput(cmd *cobra.Command) (intgraphql.NotificationEntityInput, error) {
	input := intgraphql.NotificationEntityInput{}
	set := 0

	if v, _ := cmd.Flags().GetString("issue"); v != "" {
		input.IssueID = &v
		set++
	}
	if v, _ := cmd.Flags().GetString("project"); v != "" {
		input.ProjectID = &v
		set++
	}
	if v, _ := cmd.Flags().GetString("initiative"); v != "" {
		input.InitiativeID = &v
		set++
	}
	if v, _ := cmd.Flags().GetString("notification"); v != "" {
		input.ID = &v
		set++
	}

	if set == 0 {
		return input, fmt.Errorf("one of --issue, --project, --initiative, or --notification is required")
	}
	if set > 1 {
		return input, fmt.Errorf("only one of --issue, --project, --initiative, or --notification may be specified")
	}

	return input, nil
}
