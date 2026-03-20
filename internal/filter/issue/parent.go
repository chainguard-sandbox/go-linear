package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyParent handles --parent flag.
func ApplyParent(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	parent, _ := cmd.Flags().GetString("parent")
	if parent == "" {
		return nil
	}

	b.Filter().Parent = &intgraphql.NullableIssueFilter{
		ID: &intgraphql.IssueIDComparator{Eq: &parent},
	}
	return nil
}
