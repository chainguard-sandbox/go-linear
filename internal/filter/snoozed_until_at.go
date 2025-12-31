package filter

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	intgraphql "github.com/chainguard-sandbox/go-linear/internal/graphql"
)

// ApplySnoozedUntilAt handles --snoozed-until-after and --snoozed-until-before flags.
func ApplySnoozedUntilAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("snoozed-until-after")
	before, _ := cmd.Flags().GetString("snoozed-until-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.SnoozedUntilAt == nil {
		f.SnoozedUntilAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid snoozed-until-after date: %w", err)
		}
		tStr := t.Format(DateFormat)
		f.SnoozedUntilAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid snoozed-until-before date: %w", err)
		}
		tStr := t.Format(DateFormat)
		f.SnoozedUntilAt.Lt = &tStr
	}
	return nil
}
