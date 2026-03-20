package team

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyName handles --name flag.
func ApplyName(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return nil
	}

	b.filter.Name = &intgraphql.StringComparator{ContainsIgnoreCase: &name}
	return nil
}
