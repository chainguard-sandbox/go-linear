package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyIsFuture handles --future flag.
func ApplyIsFuture(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	future, _ := cmd.Flags().GetBool("future")
	if !future {
		return nil
	}

	b.Filter().IsFuture = &intgraphql.BooleanComparator{Eq: &future}
	return nil
}
