package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyState handles --state flag.
func ApplyState(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	state, _ := cmd.Flags().GetString("state")
	if state == "" {
		return nil
	}

	stateID, err := b.Resolver().ResolveState(ctx, state)
	if err != nil {
		return fmt.Errorf("failed to resolve state: %w", err)
	}

	b.Filter().State = &intgraphql.WorkflowStateFilter{
		ID: &intgraphql.IDComparator{Eq: &stateID},
	}
	return nil
}
