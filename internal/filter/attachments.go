package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyAttachments handles --attachment-by and --attachment-source-type flags.
func ApplyAttachments(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	attachmentBy, _ := cmd.Flags().GetStringArray("attachment-by")
	attachmentSource, _ := cmd.Flags().GetString("attachment-source-type")

	if len(attachmentBy) == 0 && attachmentSource == "" {
		return nil
	}

	attachmentFilter := &intgraphql.AttachmentFilter{}

	if len(attachmentBy) > 0 {
		userIDs := make([]string, 0, len(attachmentBy))
		for _, user := range attachmentBy {
			userID, err := b.Resolver().ResolveUser(ctx, user)
			if err != nil {
				return fmt.Errorf("failed to resolve attachment-by user %q: %w", user, err)
			}
			userIDs = append(userIDs, userID)
		}

		if len(userIDs) == 1 {
			attachmentFilter.Creator = &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{Eq: &userIDs[0]},
			}
		} else {
			attachmentFilter.Creator = &intgraphql.NullableUserFilter{
				ID: &intgraphql.IDComparator{In: userIDs},
			}
		}
	}

	if attachmentSource != "" {
		attachmentFilter.SourceType = &intgraphql.SourceTypeComparator{
			Contains: &attachmentSource,
		}
	}

	b.Filter().Attachments = &intgraphql.AttachmentCollectionFilter{
		Some: attachmentFilter,
	}
	return nil
}
