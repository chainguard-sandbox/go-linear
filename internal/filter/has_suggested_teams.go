package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHasSuggestedTeams handles --has-suggested-teams flag.
func ApplyHasSuggestedTeams(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasSuggested, _ := cmd.Flags().GetBool("has-suggested-teams")
	if !hasSuggested {
		return nil
	}

	trueVal := true
	b.Filter().HasSuggestedTeams = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
