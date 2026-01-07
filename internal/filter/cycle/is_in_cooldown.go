package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyIsInCooldown handles --in-cooldown flag.
func ApplyIsInCooldown(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	inCooldown, _ := cmd.Flags().GetBool("in-cooldown")
	if !inCooldown {
		return nil
	}

	b.Filter().IsInCooldown = &intgraphql.BooleanComparator{Eq: &inCooldown}
	return nil
}
