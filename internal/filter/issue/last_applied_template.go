package issue

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyLastAppliedTemplate handles --last-applied-template flag.
func ApplyLastAppliedTemplate(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	template, _ := cmd.Flags().GetString("last-applied-template")
	if template == "" {
		return nil
	}

	b.Filter().LastAppliedTemplate = &intgraphql.NullableTemplateFilter{
		ID: &intgraphql.IDComparator{Eq: &template},
	}
	return nil
}
