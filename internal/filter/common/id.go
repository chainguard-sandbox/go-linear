package common

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyID handles --id flag.
// Works with any filter builder that implements IDFilterable.
func ApplyID[T IDFilterable](ctx context.Context, cmd *cobra.Command, b T) error {
	id, _ := cmd.Flags().GetString("id")
	if id == "" {
		return nil
	}

	b.SetID(&intgraphql.IDComparator{Eq: &id})
	return nil
}
