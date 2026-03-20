package user

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyDisplayName handles --display-name flag.
func ApplyDisplayName(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	displayName, _ := cmd.Flags().GetString("display-name")
	if displayName == "" {
		return nil
	}

	b.filter.DisplayName = &intgraphql.StringComparator{ContainsIgnoreCase: &displayName}
	return nil
}
