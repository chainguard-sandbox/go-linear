package project

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// ApplyCompletedAt handles --completed-after and --completed-before flags.
func ApplyCompletedAt(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error {
	after, _ := cmd.Flags().GetString("completed-after")
	before, _ := cmd.Flags().GetString("completed-before")

	if after == "" && before == "" {
		return nil
	}

	comp := b.CompletedAtComparator()

	if after != "" {
		t, err := b.Parser().Parse(after)
		if err != nil {
			return fmt.Errorf("invalid --completed-after: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Gte = &formatted
	}

	if before != "" {
		t, err := b.Parser().Parse(before)
		if err != nil {
			return fmt.Errorf("invalid --completed-before: %w", err)
		}
		formatted := t.Format("2006-01-02T15:04:05.000Z")
		comp.Lte = &formatted
	}

	return nil
}
