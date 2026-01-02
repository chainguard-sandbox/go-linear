package project

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyHealth handles --health flag.
// Valid values: onTrack, atRisk, offTrack
func ApplyHealth(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	health, _ := cmd.Flags().GetString("health")
	if health == "" {
		return nil
	}

	b.Filter().Health = &intgraphql.StringComparator{Eq: &health}
	return nil
}
