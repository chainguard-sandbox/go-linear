package attachment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyTitle handles --title flag (case-insensitive contains).
func ApplyTitle(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	title, _ := cmd.Flags().GetString("title")
	if title == "" {
		return nil
	}

	b.Filter().Title = &intgraphql.StringComparator{ContainsIgnoreCase: &title}
	return nil
}
