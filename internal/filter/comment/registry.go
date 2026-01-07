package comment

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/internal/filter/common"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters is the registry of all comment filter functions.
// Uses shared filters from common package where applicable.
var Filters = map[string]FilterFunc{
	// Date filters (shared)
	"created-at": common.ApplyCreatedAt[*FilterBuilder],
	"updated-at": common.ApplyUpdatedAt[*FilterBuilder],

	// Entity filters (id is shared)
	"id":      common.ApplyID[*FilterBuilder],
	"creator": ApplyCreator,
	"issue":   ApplyIssue,

	// Text filters (comment-specific)
	"body": ApplyBody,
}
