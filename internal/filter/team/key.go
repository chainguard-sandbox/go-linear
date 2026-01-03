package team

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyKey handles --key flag.
func ApplyKey(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	key, _ := cmd.Flags().GetString("key")
	if key == "" {
		return nil
	}

	b.filter.Key = &intgraphql.StringComparator{EqIgnoreCase: &key}
	return nil
}
