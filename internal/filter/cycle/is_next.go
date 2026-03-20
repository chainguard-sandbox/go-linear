package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyIsNext handles --next flag.
func ApplyIsNext(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	next, _ := cmd.Flags().GetBool("next")
	if !next {
		return nil
	}

	b.Filter().IsNext = &intgraphql.BooleanComparator{Eq: &next}
	return nil
}
