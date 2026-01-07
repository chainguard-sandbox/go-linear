package cycle

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter"
)

// ApplyStartsAt handles --starts-after and --starts-before flags.
func ApplyStartsAt(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("starts-after")
	before, _ := cmd.Flags().GetString("starts-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.StartsAtComparator()
	dr := &filter.DateRange{Gte: comp.Gte, Lt: comp.Lt}

	if err := filter.ApplyDateRange(b.Parser(), dr, after, before, "starts"); err != nil {
		return err
	}

	comp.Gte = dr.Gte
	comp.Lt = dr.Lt
	return nil
}
