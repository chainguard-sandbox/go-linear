package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyAssignee handles --assignee flag.
func ApplyAssignee(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	assignee, _ := cmd.Flags().GetString("assignee")
	if assignee == "" {
		return nil
	}

	userID, err := b.Resolver().ResolveUser(ctx, assignee)
	if err != nil {
		return fmt.Errorf("failed to resolve assignee: %w", err)
	}

	b.Filter().Assignee = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{
			Eq: &userID,
		},
	}
	return nil
}
