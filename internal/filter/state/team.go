package state

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

	// If resolver is available, resolve team to ID
	if b.Resolver() != nil {
		teamID, err := b.Resolver().ResolveTeam(ctx, team)
		if err != nil {
			return fmt.Errorf("invalid --team: %w", err)
		}
		b.filter.Team = &intgraphql.TeamFilter{
			ID: &intgraphql.IDComparator{Eq: &teamID},
		}
		return nil
	}

	// Without resolver, try as name/key
	b.filter.Team = &intgraphql.TeamFilter{
		Or: []*intgraphql.TeamFilter{
			{Name: &intgraphql.StringComparator{EqIgnoreCase: &team}},
			{Key: &intgraphql.StringComparator{EqIgnoreCase: &team}},
		},
	}
	return nil
}
