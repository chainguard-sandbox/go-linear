package project

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHasBlockedBy handles --has-blocked-by flag.
func ApplyHasBlockedBy(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	hasBlockedBy, _ := cmd.Flags().GetBool("has-blocked-by")
	if !hasBlockedBy {
		return nil
	}

	exists := true
	b.Filter().HasBlockedByRelations = &intgraphql.RelationExistsComparator{Eq: &exists}
	return nil
}

// ApplyHasBlocking handles --has-blocking flag.
func ApplyHasBlocking(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	hasBlocking, _ := cmd.Flags().GetBool("has-blocking")
	if !hasBlocking {
		return nil
	}

	exists := true
	b.Filter().HasBlockingRelations = &intgraphql.RelationExistsComparator{Eq: &exists}
	return nil
}

// ApplyHasRelated handles --has-related flag.
func ApplyHasRelated(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	hasRelated, _ := cmd.Flags().GetBool("has-related")
	if !hasRelated {
		return nil
	}

	exists := true
	b.Filter().HasRelatedRelations = &intgraphql.RelationExistsComparator{Eq: &exists}
	return nil
}
