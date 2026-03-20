package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySubscribers handles --subscriber flag.
func ApplySubscribers(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	subscribers, _ := cmd.Flags().GetStringArray("subscriber")
	if len(subscribers) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(subscribers))
	for _, user := range subscribers {
		userID, err := b.Resolver().ResolveUser(ctx, user)
		if err != nil {
			return fmt.Errorf("failed to resolve subscriber %q: %w", user, err)
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 1 {
		b.Filter().Subscribers = &intgraphql.UserCollectionFilter{
			Some: &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			},
		}
	} else {
		b.Filter().Subscribers = &intgraphql.UserCollectionFilter{
			Some: &intgraphql.UserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			},
		}
	}
	return nil
}
