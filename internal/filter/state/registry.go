package state

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters maps filter names to their apply functions.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (shared)
	"created-at": common.ApplyCreatedAt[*FilterBuilder],
	"updated-at": common.ApplyUpdatedAt[*FilterBuilder],

	// Entity filters
	"id":   common.ApplyID[*FilterBuilder],
	"team": ApplyTeam,

	// Text filters (state-specific)
	"name":        ApplyName,
	"description": ApplyDescription,
	"type":        ApplyType,
}
