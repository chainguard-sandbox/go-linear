package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyEstimate handles --estimate flag.
func ApplyEstimate(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	if cmd.Flags().Lookup("estimate") == nil {
		return nil
	}
	estimate, _ := cmd.Flags().GetInt("estimate")
	if estimate < 0 {
		return nil
	}

	e := float64(estimate)
	b.Filter().Estimate = &intgraphql.EstimateComparator{Eq: &e}
	return nil
}
