package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHasBlockedByRelations handles --has-blocked-by flag.
func ApplyHasBlockedByRelations(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasBlocked, _ := cmd.Flags().GetBool("has-blocked-by")
	if !hasBlocked {
		return nil
	}

	trueVal := true
	b.Filter().HasBlockedByRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
