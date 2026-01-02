package attachment

import (
	"context"

	"github.com/spf13/cobra"
)

// FilterFunc applies a filter based on command flags.
type FilterFunc func(ctx context.Context, cmd *cobra.Command, b *FilterBuilder) error

// Filters maps filter names to their apply functions.
// Each filter is in a separate file for maintainability.
var Filters = map[string]FilterFunc{
	// Date filters
	"created-at": ApplyCreatedAt,
	"updated-at": ApplyUpdatedAt,

	// Entity filters
	"id":          ApplyID,
	"creator":     ApplyCreator,
	"source-type": ApplySourceType,

	// Text filters
	"title":    ApplyTitle,
	"subtitle": ApplySubtitle,
	"url":      ApplyURL,
}
