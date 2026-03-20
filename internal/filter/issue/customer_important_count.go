package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyCustomerImportantCount handles --customer-important-count flag.
func ApplyCustomerImportantCount(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	if cmd.Flags().Lookup("customer-important-count") == nil {
		return nil
	}
	count, _ := cmd.Flags().GetInt("customer-important-count")
	if count < 0 {
		return nil
	}

	c := float64(count)
	b.Filter().CustomerImportantCount = &intgraphql.NumberComparator{Eq: &c}
	return nil
}
