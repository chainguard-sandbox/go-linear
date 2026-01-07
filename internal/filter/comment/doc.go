// Package comment provides filtering for Linear comments.
//
// The package implements a filter builder pattern where each filter type
// (date, entity, text) is handled by a separate function that populates
// the underlying CommentFilter struct.
//
// Available filters:
//   - Date: created-at, updated-at
//   - Entity: id, creator, issue
//   - Text: body
package comment
