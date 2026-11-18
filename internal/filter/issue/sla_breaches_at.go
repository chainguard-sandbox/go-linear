package issue

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter"
	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// ApplySLABreachesAt handles --sla-breaches-after and --sla-breaches-before flags.
//
// Relative values point forward ("7d" = 7 days from now) because breach
// times are deadlines: --sla-breaches-before=7d selects issues breaching
// within the next week, including ones already breached. Absolute dates may be
// in the past or future.
func ApplySLABreachesAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("sla-breaches-after")
	before, _ := cmd.Flags().GetString("sla-breaches-before")

	if after == "" && before == "" {
		return nil
	}

	f := b.Filter()
	if f.SLABreachesAt == nil {
		f.SLABreachesAt = &intgraphql.NullableDateComparator{}
	}

	if after != "" {
		t, err := b.Parser().ParseDeadline(after)
		if err != nil {
			return fmt.Errorf("invalid sla-breaches-after date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.SLABreachesAt.Gte = &tStr
	}
	if before != "" {
		t, err := b.Parser().ParseDeadline(before)
		if err != nil {
			return fmt.Errorf("invalid sla-breaches-before date: %w", err)
		}
		tStr := t.Format(filter.DateFormat)
		f.SLABreachesAt.Lt = &tStr
	}
	return nil
}
