package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyProjectMilestone handles --project-milestone flag.
func ApplyProjectMilestone(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	milestone, _ := cmd.Flags().GetString("project-milestone")
	if milestone == "" {
		return nil
	}

	b.Filter().ProjectMilestone = &intgraphql.NullableProjectMilestoneFilter{
		ID: &intgraphql.IDComparator{Eq: &milestone},
	}
	return nil
}
