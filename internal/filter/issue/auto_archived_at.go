package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyAutoArchivedAt handles --auto-archived-after and --auto-archived-before flags.
func ApplyAutoArchivedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("auto-archived-after")
	before, _ := cmd.Flags().GetString("auto-archived-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.AutoArchivedAt == nil {
		f.AutoArchivedAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid auto-archived-after date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AutoArchivedAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid auto-archived-before date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AutoArchivedAt.Lt = &tStr
	}
	return nil
}
