package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyDelegate handles --delegate flag.
func ApplyDelegate(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	delegate, _ := cmd.Flags().GetString("delegate")
	if delegate == "" {
		return nil
	}

	userID, err := b.Resolver().ResolveUser(ctx, delegate)
	if err != nil {
		return fmt.Errorf("failed to resolve delegate: %w", err)
	}

	b.Filter().Delegate = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
