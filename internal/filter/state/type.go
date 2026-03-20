package state

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyType handles --type flag.
// Valid types: triage, backlog, unstarted, started, completed, canceled
func ApplyType(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	stateType, _ := cmd.Flags().GetString("type")
	if stateType == "" {
		return nil
	}

	b.filter.Type = &intgraphql.StringComparator{EqIgnoreCase: &stateType}
	return nil
}
