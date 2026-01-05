package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyCycle handles --cycle flag.
func ApplyCycle(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	cycle, _ := cmd.Flags().GetString("cycle")
	if cycle == "" {
		return nil
	}

	b.Filter().Cycle = &intgraphql.NullableCycleFilter{
		ID: &intgraphql.IDComparator{Eq: &cycle},
	}
	return nil
}
