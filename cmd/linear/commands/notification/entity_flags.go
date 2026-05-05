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
	cmd.Flags().String("project", "", "Project name or UUID (deprecated by Linear API; may stop working server-side)")
	cmd.Flags().String("initiative", "", "Initiative name or UUID")
	cmd.Flags().String("notification", "", "Notification ID (UUID)")
}

// buildEntityInput constructs a NotificationEntityInput from flags, resolving
// human-readable identifiers (issue keys, project/initiative names) to UUIDs.
func buildEntityInput(cmd *cobra.Command, ctx context.Context, res *resolver.Resolver) (intgraphql.NotificationEntityInput, error) {
	input := intgraphql.NotificationEntityInput{}

	issueVal, _ := cmd.Flags().GetString("issue")
	projectVal, _ := cmd.Flags().GetString("project")
	initiativeVal, _ := cmd.Flags().GetString("initiative")
	notifVal, _ := cmd.Flags().GetString("notification")

	set := 0
	if issueVal != "" {
		set++
	}
	if projectVal != "" {
		set++
	}
	if initiativeVal != "" {
		set++
	}
	if notifVal != "" {
		set++
	}

	if set == 0 {
		return input, fmt.Errorf("one of --issue, --project, --initiative, or --notification is required")
	}
	if set > 1 {
		return input, fmt.Errorf("only one of --issue, --project, --initiative, or --notification may be specified")
	}

	switch {
	case issueVal != "":
		id, err := res.ResolveIssue(ctx, issueVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve issue: %w", err)
		}
		input.IssueID = &id
	case projectVal != "":
		id, err := res.ResolveProject(ctx, projectVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve project: %w", err)
		}
		input.ProjectID = &id
	case initiativeVal != "":
		id, err := res.ResolveInitiative(ctx, initiativeVal)
		if err != nil {
			return input, fmt.Errorf("failed to resolve initiative: %w", err)
		}
		input.InitiativeID = &id
	case notifVal != "":
		input.ID = &notifVal
	}

	return input, nil
}
