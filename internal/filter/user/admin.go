package user

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyAdmin handles --admin flag.
func ApplyAdmin(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	if !cmd.Flags().Changed("admin") {
		return nil
	}

	admin, _ := cmd.Flags().GetBool("admin")
	b.filter.Admin = &intgraphql.BooleanComparator{Eq: &admin}
	return nil
}
