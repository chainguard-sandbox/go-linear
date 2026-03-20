package issue

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/internal/filter"
)

// ApplyStartedAt handles --started-after and --started-before flags.
func ApplyStartedAt(ctx context.Context, cmd *cobra.Command, b *IssueFilterBuilder) error {
	after, _ := cmd.Flags().GetString("started-after")
	before, _ := cmd.Flags().GetString("started-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.StartedAtComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "started"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
