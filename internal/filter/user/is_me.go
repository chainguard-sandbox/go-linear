package user

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyIsMe handles --is-me flag.
func ApplyIsMe(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	if !cmd.Flags().Changed("is-me") {
		return nil
	}

	isMe, _ := cmd.Flags().GetBool("is-me")
	b.filter.IsMe = &intgraphql.BooleanComparator{Eq: &isMe}
	return nil
}
