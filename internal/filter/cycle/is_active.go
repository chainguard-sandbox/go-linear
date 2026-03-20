package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyIsActive handles --active flag.
func ApplyIsActive(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	active, _ := cmd.Flags().GetBool("active")
	if !active {
		return nil
	}

	b.Filter().IsActive = &intgraphql.BooleanComparator{Eq: &active}
	return nil
}
