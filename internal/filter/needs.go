package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyNeeds handles --has-needs flag.
func ApplyNeeds(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasNeeds, _ := cmd.Flags().GetBool("has-needs")
	if !hasNeeds {
		return nil
	}

	zero := float64(0)
	b.Filter().Needs = &intgraphql.CustomerNeedCollectionFilter{
		Length: &intgraphql.NumberComparator{Gt: &zero},
	}
	return nil
}
