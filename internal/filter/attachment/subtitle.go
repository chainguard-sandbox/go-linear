package attachment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySubtitle handles --subtitle flag (case-insensitive contains).
func ApplySubtitle(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	subtitle, _ := cmd.Flags().GetString("subtitle")
	if subtitle == "" {
		return nil
	}

	b.Filter().Subtitle = &intgraphql.NullableStringComparator{ContainsIgnoreCase: &subtitle}
	return nil
}
