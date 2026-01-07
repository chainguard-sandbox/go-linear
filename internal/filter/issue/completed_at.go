package issue

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
)

// ApplyCompletedAt handles --completed-after and --completed-before flags.
func ApplyCompletedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("completed-after")
	before, _ := cmd.Flags().GetString("completed-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.CompletedAtComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "completed"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
