package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHasSuggestedAssignees handles --has-suggested-assignees flag.
func ApplyHasSuggestedAssignees(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasSuggested, _ := cmd.Flags().GetBool("has-suggested-assignees")
	if !hasSuggested {
		return nil
	}

	trueVal := true
	b.Filter().HasSuggestedAssignees = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
