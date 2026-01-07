package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHasSuggestedLabels handles --has-suggested-labels flag.
func ApplyHasSuggestedLabels(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasSuggested, _ := cmd.Flags().GetBool("has-suggested-labels")
	if !hasSuggested {
		return nil
	}

	trueVal := true
	b.Filter().HasSuggestedLabels = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
