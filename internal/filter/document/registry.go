package document

// Filters maps filter names to their apply functions.
var Filters = map[string]FilterFunc{
	// Date filters
	"created-at": ApplyCreatedAt,
	"updated-at": ApplyUpdatedAt,

	// Entity filters
	"id":         ApplyID,
	"creator":    ApplyCreator,
	"initiative": ApplyInitiative,
	"project":    ApplyProject,
	"issue":      ApplyIssue,

	// Text filters
	"title":   ApplyTitle,
	"slug-id": ApplySlugID,
}
