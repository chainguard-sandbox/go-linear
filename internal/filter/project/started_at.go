package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyStartedAt handles --started-after and --started-before flags.
func ApplyStartedAt(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("started-after")
	before, _ := cmd.Flags().GetString("started-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.StartedAtComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --started-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --started-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
