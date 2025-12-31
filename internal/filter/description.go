package filter

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyDescription handles --description flag.
func ApplyDescription(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	description, _ := cmd.Flags().GetString("description")
	if description == "" {
		return nil
	}

	b.Filter().Description = &intgraphql.NullableStringComparator{
		Contains: &description,
	}
	return nil
}
