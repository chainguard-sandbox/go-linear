package initiative

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyOwner handles --owner flag.
func ApplyOwner(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	owner, _ := cmd.Flags().GetString("owner")
	if owner == "" {
		return nil
	}

	userID, err := b.Resolver().ResolveUser(ctx, owner)
	if err != nil {
		return fmt.Errorf("failed to resolve owner: %w", err)
	}

	b.Filter().Owner = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
