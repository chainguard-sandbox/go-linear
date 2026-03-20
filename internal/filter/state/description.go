package state

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyDescription handles --description flag.
func ApplyDescription(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	desc, _ := cmd.Flags().GetString("description")
	if desc == "" {
		return nil
	}

	b.filter.Description = &intgraphql.StringComparator{ContainsIgnoreCase: &desc}
	return nil
}
