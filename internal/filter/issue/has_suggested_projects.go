package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyHasSuggestedProjects handles --has-suggested-projects flag.
func ApplyHasSuggestedProjects(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasSuggested, _ := cmd.Flags().GetBool("has-suggested-projects")
	if !hasSuggested {
		return nil
	}

	trueVal := true
	b.Filter().HasSuggestedProjects = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
