package initiative

// Filters maps filter names to their apply functions.
var Filters = map[string]FilterFunc{
	// Date filters
	"created-at":  ApplyCreatedAt,
	"target-date": ApplyTargetDate,

	// Entity filters
	"id":      ApplyID,
	"creator": ApplyCreator,
	"owner":   ApplyOwner,

	// Text filters
	"name":    ApplyName,
	"slug-id": ApplySlugID,

	// State filters
	"health": ApplyHealth,
	"status": ApplyStatus,
}
