package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyComments handles --comment-by and --comment-contains flags.
func ApplyComments(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	commentBy, _ := cmd.Flags().GetStringArray("comment-by")
	commentContains, _ := cmd.Flags().GetString("comment-contains")

	if len(commentBy) == 0 && commentContains == "" {
		return nil
	}

	commentFilter := &intgraphql.CommentFilter{}

	if len(commentBy) > 0 {
		userIDs := make([]string, 0, len(commentBy))
		for _, user := range commentBy {
			userID, err := b.Resolver().ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve comment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		if len(userIDs) == 1 {
			commentFilter.User = &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			}
		} else {
			commentFilter.User = &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			}
		}
	}

	if commentContains != "" {
		commentFilter.Body = &intgraphql.StringComparator{
			Contains: &commentContains,
		}
	}

	b.Filter().Comments = &intgraphql.CommentCollectionFilter{
		Some: commentFilter,
	}
	return nil
}
