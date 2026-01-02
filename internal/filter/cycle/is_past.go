package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyIsPast handles --past flag.
func ApplyIsPast(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	past, _ := cmd.Flags().GetBool("past")
	if !past {
		return nil
	}

	b.Filter().IsPast = &intgraphql.BooleanComparator{Eq: &past}
	return nil
}
