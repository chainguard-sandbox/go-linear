package document

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySlugID handles --slug-id flag.
func ApplySlugID(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	slugID, _ := cmd.Flags().GetString("slug-id")
	if slugID == "" {
		return nil
	}

	b.Filter().SlugID = &intgraphql.StringComparator{Eq: &slugID}
	return nil
}
