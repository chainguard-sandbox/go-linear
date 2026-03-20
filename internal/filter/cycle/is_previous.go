package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyIsPrevious handles --previous flag.
func ApplyIsPrevious(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	previous, _ := cmd.Flags().GetBool("previous")
	if !previous {
		return nil
	}

	b.Filter().IsPrevious = &intgraphql.BooleanComparator{Eq: &previous}
	return nil
}
