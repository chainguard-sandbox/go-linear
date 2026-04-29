package notification

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
)

// addEntityFlags adds the common entity filter flags for bulk notification operations.
func addEntityFlags(cmd *cobra.Command) {
	cmd.Flags().String("issue", "", "Issue identifier or UUID (e.g., ENG-123)")
	cmd.Flags().String("project", "", "Project name or UUID")
	cmd.Flags().String("initiative", "", "Initiative name or UUID")
	cmd.Flags().String("notification", "", "Notification ID (UUID)")
}

// buildEntityInput constructs a NotificationEntityInput from flags, resolving
// human-readable identifiers (issue keys, project/initiative names) to UUIDs.
func buildEntityInput(cmd *cobra.Command, ctx context.Context, res *resolver.Resolver) (intgraphql.NotificationEntityInput, error) {
	input := intgraphql.NotificationEntityInput{}
	set := 0

	if v, _ := cmd.Flags().GetString("issue"); v != "" {
		id, err := res.ResolveIssue(ctx, v)
		if err != nil {
			return input, fmt.Errorf("failed to resolve issue: %w", err)
		}
		input.IssueID = &id
		set++
	}
	if v, _ := cmd.Flags().GetString("project"); v != "" {
		id, err := res.ResolveProject(ctx, v)
		if err != nil {
			return input, fmt.Errorf("failed to resolve project: %w", err)
		}
		input.ProjectID = &id
		set++
	}
	if v, _ := cmd.Flags().GetString("initiative"); v != "" {
		id, err := res.ResolveInitiative(ctx, v)
		if err != nil {
			return input, fmt.Errorf("failed to resolve initiative: %w", err)
		}
		input.InitiativeID = &id
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
