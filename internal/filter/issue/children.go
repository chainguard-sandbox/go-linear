package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyChildren handles --has-children flag.
func ApplyChildren(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasChildren, _ := cmd.Flags().GetBool("has-children")
	if !hasChildren {
		return nil
	}

	zero := float64(0)
	b.Filter().Children = &intgraphql.IssueCollectionFilter{
		Length: &intgraphql.NumberComparator{
			Gt: &zero,
		},
	}
	return nil
}
