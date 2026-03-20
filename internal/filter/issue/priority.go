package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyPriority handles --priority flag.
func ApplyPriority(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	if cmd.Flags().Lookup("priority") == nil {
		return nil
	}
	priority, _ := cmd.Flags().GetInt("priority")
	if priority < 0 {
		return nil
	}

	p := float64(priority)
	b.Filter().Priority = &intgraphql.NullableNumberComparator{Eq: &p}
	return nil
}
