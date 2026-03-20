package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyHasDuplicateRelations handles --has-duplicate flag.
func ApplyHasDuplicateRelations(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	hasDupe, _ := cmd.Flags().GetBool("has-duplicate")
	if !hasDupe {
		return nil
	}

	trueVal := true
	b.Filter().HasDuplicateRelations = &intgraphql.RelationExistsComparator{Eq: &trueVal}
	return nil
}
