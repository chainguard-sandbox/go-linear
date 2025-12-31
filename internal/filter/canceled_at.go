package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplyCanceledAt handles --canceled-after and --canceled-before flags.
func ApplyCanceledAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("canceled-after")
	before, _ := cmd.Flags().GetString("canceled-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.CanceledAt == nil {
		f.CanceledAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid canceled-after date: %w", err)
		}
		tStr := t.Format(DateFormat)
		f.CanceledAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid canceled-before date: %w", err)
		}
		tStr := t.Format(DateFormat)
		f.CanceledAt.Lt = &tStr
	}
	return nil
}
