package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyAddedToCyclePeriod handles --added-to-cycle-period flag.
func ApplyAddedToCyclePeriod(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	period, _ := cmd.Flags().GetString("added-to-cycle-period")
	if period == "" {
		return nil
	}

	b.Filter().AddedToCyclePeriod = &intgraphql.CyclePeriodComparator{
		Eq: (*intgraphql.CyclePeriod)(&period),
	}
	return nil
}
