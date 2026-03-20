package user

import (
	"context"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyEmail handles --email flag.
func ApplyEmail(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	email, _ := cmd.Flags().GetString("email")
	if email == "" {
		return nil
	}

	b.filter.Email = &intgraphql.StringComparator{ContainsIgnoreCase: &email}
	return nil
}
