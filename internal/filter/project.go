package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyProject handles --project flag.
func ApplyProject(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	project, _ := cmd.Flags().GetString("project")
	if project == "" {
		return nil
	}

	b.Filter().Project = &intgraphql.NullableProjectFilter{
		ID: &intgraphql.IDComparator{Eq: &project},
	}
	return nil
}
