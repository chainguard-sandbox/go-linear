package notification

import (
	"context"
	"fmt"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
	"github.com/chainguard-sandbox/go-linear/v2/internal/resolver"
	"github.com/spf13/cobra"
)

// addEntityFlags adds the common entity filter flags for bulk notification operations.
func addEntityFlags(cmd *cobra.Command) {
	cmd.Flags().String("issue", "", "Issue identifier or UUID (e.g., ENG-123)")
	cmd.Flags().String("project", "", "Project name or UUID (deprecated by Linear API; may stop working server-side)")
	cmd.Flags().String("initiative", "", "Initiative name or UUID")
	cmd.Flags().String("notification", "", "Notification ID (UUID)")
}

// buildEntityInput constructs a NotificationEntityInput from flags, resolving
// human-readable identifiers (issue keys, project/initiative names) to UUIDs.
func buildEntityInput(cmd *cobra.Command, ctx context.Context, res *resolver.Resolver) (intgraphql.NotificationEntityInput, error) {
	input := intgraphql.NotificationEntityInput{}
	set := 0

	issueVal, err := cmd.Flags().GetString("issue")
	if err != nil {
		return input, err
	}
	if issueVal != "" {
		id, err := res.ResolveIssue(ctx, issueVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve issue: %w", err)
		}
		input.IssueID = &id
		set++
	}

	projectVal, err := cmd.Flags().GetString("project")
	if err != nil {
		return input, err
	}
	if projectVal != "" {
		id, err := res.ResolveProject(ctx, projectVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve project: %w", err)
		}
		input.ProjectID = &id
		set++
	}

	initiativeVal, err := cmd.Flags().GetString("initiative")
	if err != nil {
		return input, err
	}
	if initiativeVal != "" {
		id, err := res.ResolveInitiative(ctx, initiativeVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve initiative: %w", err)
		}
		input.InitiativeID = &id
		set++
	}

	notifVal, err := cmd.Flags().GetString("notification")
	if err != nil {
		return input, err
	}
	if notifVal != "" {
		input.ID = &notifVal
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
