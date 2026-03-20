package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplyAutoClosedAt handles --auto-closed-after and --auto-closed-before flags.
func ApplyAutoClosedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("auto-closed-after")
	before, _ := cmd.Flags().GetString("auto-closed-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.AutoClosedAt == nil {
		f.AutoClosedAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid auto-closed-after date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AutoClosedAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid auto-closed-before date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AutoClosedAt.Lt = &tStr
	}
	return nil
}
