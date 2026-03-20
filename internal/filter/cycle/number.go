package cycle

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyNumber handles --number flag.
func ApplyNumber(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	number, _ := cmd.Flags().GetInt("number")
	if number < 0 {
		return nil
	}

	num := float64(number)
	b.Filter().Number = &intgraphql.NumberComparator{Eq: &num}
	return nil
}
