package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyLabels handles --label flag.
func ApplyLabels(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	labels, _ := cmd.Flags().GetStringArray("label")
	if len(labels) == 0 {
		return nil
	}

	labelIDs := make([]string, 0, len(labels))
	for _, label := range labels {
		labelID, err := b.Resolver().ResolveLabel(ctx, label)
		if err != nil {
			return fmt.Errorf("failed to resolve label %q: %w", label, err)
		}
		labelIDs = append(labelIDs, labelID)
	}

	if len(labelIDs) > 0 {
		filters := make([]*intgraphql.IssueLabelFilter, len(labelIDs))
		for i, id := range labelIDs {
			idCopy := id
			filters[i] = &intgraphql.IssueLabelFilter{
				ID: &intgraphql.IDComparator{Eq: &idCopy},
			}
		}
		b.Filter().Labels = &intgraphql.IssueLabelCollectionFilter{
			Some: &intgraphql.IssueLabelFilter{Or: filters},
		}
	}
	return nil
}
