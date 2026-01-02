package label

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyTeam handles --team flag.
func ApplyTeam(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	team, _ := cmd.Flags().GetString("team")
	if team == "" {
		return nil
	}

	teamID, err := b.Resolver().ResolveTeam(ctx, team)
	if err != nil {
		return fmt.Errorf("resolving team: %w", err)
	}

	b.Filter().Team = &intgraphql.NullableTeamFilter{
		ID: &intgraphql.IDComparator{Eq: &teamID},
	}
	return nil
}
