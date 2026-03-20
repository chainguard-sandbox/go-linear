package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySnoozedBy handles --snoozed-by flag.
func ApplySnoozedBy(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	snoozedBy, _ := cmd.Flags().GetString("snoozed-by")
	if snoozedBy == "" {
		return nil
	}

	userID, err := b.Resolver().ResolveUser(ctx, snoozedBy)
	if err != nil {
		return fmt.Errorf("failed to resolve snoozed-by: %w", err)
	}

	b.Filter().SnoozedBy = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
