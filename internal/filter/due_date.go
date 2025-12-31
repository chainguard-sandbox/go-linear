package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// TimelessDateFormat is used for due dates (date only, no time).
const TimelessDateFormat = "2006-01-02"

// ApplyDueDate handles --due-after and --due-before flags.
func ApplyDueDate(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("due-after")
	before, _ := cmd.Flags().GetString("due-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.DueDate == nil {
		f.DueDate = &intgraphql.NullableTimelessDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid due-after date: %w", err)
		}
		tStr := t.Format(TimelessDateFormat)
		f.DueDate.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid due-before date: %w", err)
		}
		tStr := t.Format(TimelessDateFormat)
		f.DueDate.Lt = &tStr
	}
	return nil
}
