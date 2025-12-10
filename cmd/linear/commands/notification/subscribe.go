package notification

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/formatter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// NewSubscribeCommand creates the notification subscribe command.
func NewSubscribeCommand(clientFactory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribe",
		Short: "Subscribe to notifications for a project, cycle, team, or user",
		Long: `Subscribe to receive notifications for updates to a project, cycle, team, or user.

This operation creates new subscription data and is safe to execute.
Get notified when the resource is updated or has new activity.

Parameters (mutually exclusive - specify exactly one):
  --project: Project name or UUID
  --cycle: Cycle UUID
  --team: Team name or UUID
  --user: User name, email, or UUID

Examples:
  # Subscribe to project updates
  linear notification subscribe --project="Platform Redesign"

  # Subscribe to team updates
  linear notification subscribe --team=Engineering

  # Subscribe to cycle updates
  linear notification subscribe --cycle=<cycle-uuid> --output=json

TIP: Subscriptions help track important resources without manual checking

Common Errors:
  - "must specify exactly one resource": Provide one of the resource flags

Related Commands:
  - linear notification unsubscribe - Remove subscription
  - linear notification update - Manage notifications
  - linear notification archive - Archive notifications`,
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
	cmd.Flags().String("user", "", "User name, email, or UUID")
	cmd.Flags().StringP("output", "o", "table", "Output format: json|table")

	return cmd
}

func runSubscribe(cmd *cobra.Command, client *linear.Client) error {
	ctx := context.Background()

	projectID, _ := cmd.Flags().GetString("project")
	cycleID, _ := cmd.Flags().GetString("cycle")
	teamID, _ := cmd.Flags().GetString("team")
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
	if userID != "" {
		resourceCount++
		resourceType = "user"
	}

	if resourceCount == 0 {
		return fmt.Errorf("must specify one of: --project, --cycle, --team, or --user")
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
