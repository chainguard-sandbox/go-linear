package document

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyIssue handles --issue flag.
func ApplyIssue(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	issue, _ := cmd.Flags().GetString("issue")
	if issue == "" {
		return nil
	}

	issueID, err := b.Resolver().ResolveIssue(ctx, issue)
	if err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	b.Filter().Issue = &intgraphql.IssueFilter{
		ID: &intgraphql.IssueIDComparator{Eq: &issueID},
	}
	return nil
}
