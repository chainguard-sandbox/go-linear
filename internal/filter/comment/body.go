package comment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyBody handles --body flag.
func ApplyBody(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	body, _ := cmd.Flags().GetString("body")
	if body == "" {
		return nil
	}

	b.Filter().Body = &intgraphql.StringComparator{ContainsIgnoreCase: &body}
	return nil
}
