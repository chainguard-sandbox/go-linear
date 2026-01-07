// Package project provides filtering for Linear projects.
//
// The package implements a filter builder pattern where each filter type
// (date, entity, text, state, relation) is handled by a separate function
// that populates the underlying ProjectFilter struct.
//
// Available filters:
//   - Date: created-at, updated-at, completed-at, canceled-at, started-at, target-date
//   - Entity: id, creator, lead
//   - Text: name, slug-id
//   - State: health
//   - Numeric: priority
//   - Relation: has-blocked-by, has-blocking, has-related
package project
