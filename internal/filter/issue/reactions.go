package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyReactions handles --has-reactions flag.
func ApplyReactions(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasReactions, _ := cmd.Flags().GetBool("has-reactions")
	if !hasReactions {
		return nil
	}

	zero := float64(0)
	b.Filter().Reactions = &intgraphql.ReactionCollectionFilter{
		Length: &intgraphql.NumberComparator{Gt: &zero},
	}
	return nil
}
