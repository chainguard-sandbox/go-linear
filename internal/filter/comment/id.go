package comment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyID handles --id flag.
func ApplyID(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		return nil
	}

	b.Filter().ID = &intgraphql.IDComparator{Eq: &id}
	return nil
}
