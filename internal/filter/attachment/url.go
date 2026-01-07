package attachment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyURL handles --url flag (case-insensitive contains).
func ApplyURL(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	url, _ := cmd.Flags().GetString("url")
	if url == "" {
		return nil
	}

	b.Filter().URL = &intgraphql.StringComparator{ContainsIgnoreCase: &url}
	return nil
}
