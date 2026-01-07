package initiative

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyStatus handles --status flag.
// Valid values: Planned, Active, Completed
func ApplyStatus(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	status, _ := cmd.Flags().GetString("status")
	if status == "" {
		return nil
	}

	b.Filter().Status = &intgraphql.StringComparator{Eq: &status}
	return nil
}
