package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyInitiative handles --initiative flag.
func ApplyInitiative(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	initiative, _ := cmd.Flags().GetString("initiative")
	if initiative == "" {
		return nil
	}

	initiativeID, err := b.Resolver().ResolveInitiative(ctx, initiative)
	if err != nil {
		return fmt.Errorf("failed to resolve initiative: %w", err)
	}

	b.Filter().Initiative = &intgraphql.InitiativeFilter{
		ID: &intgraphql.IDComparator{Eq: &initiativeID},
	}
	return nil
}
