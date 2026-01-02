package comment

import (
	"context"

	"github.com/spf13/cobra"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters is the registry of all comment filter functions.
var Filters = map[string]FilterFunc{
	// Date filters
	"created-at": ApplyCreatedAt,
	"updated-at": ApplyUpdatedAt,

	// Entity filters
	"id":      ApplyID,
	"creator": ApplyCreator,
	"issue":   ApplyIssue,

	// Text filters
	"body": ApplyBody,
}
