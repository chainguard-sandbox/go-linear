package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyNumber handles --number flag.
func ApplyNumber(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	if cmd.Flags().Lookup("number") == nil {
		return nil
	}
	number, _ := cmd.Flags().GetInt("number")
	if number < 0 {
		return nil
	}

	n := float64(number)
	b.Filter().Number = &intgraphql.NumberComparator{Eq: &n}
	return nil
}
