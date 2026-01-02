package label

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyIsGroup handles --is-group flag.
func ApplyIsGroup(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	if !cmd.Flags().Changed("is-group") {
		return nil
	}

	isGroup, _ := cmd.Flags().GetBool("is-group")
	b.Filter().IsGroup = &intgraphql.BooleanComparator{Eq: &isGroup}
	return nil
}
