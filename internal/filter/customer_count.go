package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyCustomerCount handles --customer-count flag.
func ApplyCustomerCount(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	if cmd.Flags().Lookup("customer-count") == nil {
		return nil
	}
	count, _ := cmd.Flags().GetInt("customer-count")
	if count < 0 {
		return nil
	}

	c := float64(count)
	b.Filter().CustomerCount = &intgraphql.NumberComparator{Eq: &c}
	return nil
}
