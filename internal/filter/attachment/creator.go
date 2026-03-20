package attachment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyCreator handles --creator flag.
func ApplyCreator(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	creator, _ := cmd.Flags().GetString("creator")
	if creator == "" {
		return nil
	}

	userID, err := b.Resolver().ResolveUser(ctx, creator)
	if err != nil {
		return fmt.Errorf("resolving creator: %w", err)
	}

	b.Filter().Creator = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
