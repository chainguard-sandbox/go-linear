package project

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyPriority handles --priority flag.
// Valid values: 0=none, 1=urgent, 2=high, 3=normal, 4=low
func ApplyPriority(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	priority, _ := cmd.Flags().GetInt("priority")
	if !cmd.Flags().Changed("priority") {
		return nil
	}

	p := float64(priority)
	b.Filter().Priority = &intgraphql.NullableNumberComparator{Eq: &p}
	return nil
}
