package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyHasBlockingRelations handles --has-blocking flag.
func ApplyHasBlockingRelations(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasBlocking, _ := cmd.Flags().GetBool("has-blocking")
	if !hasBlocking {
		return nil
	}

	trueVal := true
	b.Filter().HasBlockingRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
