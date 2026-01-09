package notification

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewSubscribeCommand creates the notification subscribe command.
func NewSubscribeCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to notifications for a project, cycle, team, initiative, or user",
		Long: `Subscribe to updates. Safe operation. Must specify exactly one of: --project, --cycle, --team, --initiative, --user.

Note: Linear API does not support subscribing to individual issues.
To follow an issue, subscribe to its project or use project-level notifications.

Example: go-linear notification subscribe --project=<uuid> --output=json

Related: notification_unsubscribe, notification_update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFactory()
			if err != nil {
				return err
			}
			defer client.Close()

			return runSubscribe(cmd, client)
		},
	}

	cmd.Flags().String("project", "", "Project name or UUID")
	cmd.Flags().String("cycle", "", "Cycle UUID")
	cmd.Flags().String("team", "", "Team name or UUID")
	cmd.Flags().String("initiative", "", "Initiative name or UUID")
	cmd.Flags().String("user", "", "User name, email, or UUID")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runSubscribe(cmd *cobra.Command, client *linear.Client) error {
	ctx := cmd.Context()

	projectID, _ := cmd.Flags().GetString("project")
	cycleID, _ := cmd.Flags().GetString("cycle")
	teamID, _ := cmd.Flags().GetString("team")
	initiativeID, _ := cmd.Flags().GetString("initiative")
	userID, _ := cmd.Flags().GetString("user")

	// Validate exactly one resource type specified
	resourceCount := 0
	resourceType := ""
	if projectID != "" {
		resourceCount++
		resourceType = "project"
	}
	if cycleID != "" {
		resourceCount++
		resourceType = "cycle"
	}
	if teamID != "" {
		resourceCount++
		resourceType = "team"
	}
	if initiativeID != "" {
		resourceCount++
		resourceType = "initiative"
	}
	if userID != "" {
		resourceCount++
		resourceType = "user"
	}

	if resourceCount == 0 {
		return fmt.Errorf("must specify one of: --project, --cycle, --team, --initiative, or --user")
	}
	if resourceCount > 1 {
		return fmt.Errorf("must specify exactly one resource type, not multiple")
	}

	// Build input
	input := intgraphql.NotificationSubscriptionCreateInput{}
	switch resourceType {
	case "project":
		input.ProjectID = &projectID
	case "cycle":
		input.CycleID = &cycleID
	case "team":
		input.TeamID = &teamID
	case "initiative":
		input.InitiativeID = &initiativeID
	case "user":
		input.UserID = &userID
	}

	result, err := client.NotificationSubscriptionCreate(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	switch output {
	case "json":
		return formatter.FormatJSON(cmd.OutOrStdout(), result, true)
	case "table":
		fmt.Fprintf(cmd.OutOrStdout(), "✓ Subscribed to %s updates\n", resourceType)
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", output)
	}
}
