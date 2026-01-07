package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyTargetDate handles --target-after and --target-before flags.
func ApplyTargetDate(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("target-after")
	before, _ := cmd.Flags().GetString("target-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.TargetDateComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --target-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --target-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
