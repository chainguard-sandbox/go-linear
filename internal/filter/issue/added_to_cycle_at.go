package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyAddedToCycleAt handles --added-to-cycle-after and --added-to-cycle-before flags.
func ApplyAddedToCycleAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("added-to-cycle-after")
	before, _ := cmd.Flags().GetString("added-to-cycle-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.AddedToCycleAt == nil {
		f.AddedToCycleAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid added-to-cycle-after date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AddedToCycleAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid added-to-cycle-before date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.AddedToCycleAt.Lt = &tStr
	}
	return nil
}
