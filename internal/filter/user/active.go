package user

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyActive handles --active flag.
func ApplyActive(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	if !cmd.Flags().Changed("active") {
		return nil
	}

	active, _ := cmd.Flags().GetBool("active")
	b.filter.Active = &intgraphql.BooleanComparator{Eq: &active}
	return nil
}
