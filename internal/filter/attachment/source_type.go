package attachment

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySourceType handles --source-type flag.
// Valid values: uploaded, url, github, slack
func ApplySourceType(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	sourceType, _ := cmd.Flags().GetString("source-type")
	if sourceType == "" {
		return nil
	}

	b.Filter().SourceType = &intgraphql.SourceTypeComparator{Eq: &sourceType}
	return nil
}
