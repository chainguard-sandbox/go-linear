package comment

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyIssue handles --issue flag.
func ApplyIssue(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	issue, _ := cmd.Flags().GetString("issue")
	if issue == "" {
		return nil
	}

	res := b.Resolver()
	if res == nil {
		return fmt.Errorf("resolver required for --issue filter")
	}

	issueID, err := res.ResolveIssue(ctx, issue)
	if err != nil {
		return fmt.Errorf("invalid --issue: %w", err)
	}

	b.Filter().Issue = &intgraphql.NullableIssueFilter{
		ID: &intgraphql.IssueIDComparator{Eq: &issueID},
	}
	return nil
}
