package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyTriagedAt handles --triaged-after and --triaged-before flags.
func ApplyTriagedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("triaged-after")
	before, _ := cmd.Flags().GetString("triaged-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.TriagedAt == nil {
		f.TriagedAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid triaged-after date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.TriagedAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid triaged-before date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.TriagedAt.Lt = &tStr
	}
	return nil
}
