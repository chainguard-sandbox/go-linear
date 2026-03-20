package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyHasRelatedRelations handles --has-related flag.
func ApplyHasRelatedRelations(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasRelated, _ := cmd.Flags().GetBool("has-related")
	if !hasRelated {
		return nil
	}

	trueVal := true
	b.Filter().HasRelatedRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
