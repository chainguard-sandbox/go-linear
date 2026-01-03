package team

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyPrivate handles --private flag.
func ApplyPrivate(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	if !cmd.Flags().Changed("private") {
		return nil
	}

	private, _ := cmd.Flags().GetBool("private")
	b.filter.Private = &intgraphql.BooleanComparator{Eq: &private}
	return nil
}
