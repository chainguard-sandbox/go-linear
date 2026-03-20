package comment

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

	res := b.Resolver()
	if res == nil {
		return fmt.Errorf("resolver required for --creator filter")
	}

	userID, err := res.ResolveUser(ctx, creator)
	if err != nil {
		return fmt.Errorf("invalid --creator: %w", err)
	}

	b.Filter().User = &intgraphql.UserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
