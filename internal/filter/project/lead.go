package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyLead handles --lead flag.
func ApplyLead(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	lead, _ := cmd.Flags().GetString("lead")
	if lead == "" {
		return nil
	}

	res := b.Resolver()
	if res == nil {
		return fmt.Errorf("resolver required for --lead filter")
	}

	userID, err := res.ResolveUser(ctx, lead)
	if err != nil {
		return fmt.Errorf("invalid --lead: %w", err)
	}

	b.Filter().Lead = &intgraphql.NullableUserFilter{
		ID: &intgraphql.IDComparator{Eq: &userID},
	}
	return nil
}
