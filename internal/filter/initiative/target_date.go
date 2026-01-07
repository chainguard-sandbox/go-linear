package initiative

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
)

// ApplyTargetDate handles --target-after and --target-before flags.
func ApplyTargetDate(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("target-after")
	before, _ := cmd.Flags().GetString("target-before")
	if after == "" && before == "" {
		return nil
	}

	comp := b.TargetDateComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}
	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "target"); err != nil {
		return err
	}
	comp.Gte, comp.Lt = dr.Gte, dr.Lt
	return nil
}
