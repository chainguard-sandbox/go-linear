// Package label provides filtering for Linear issue labels.
//
// The package implements 7 filter types:
//   - 2 date filters: created-at, updated-at
//   - 3 entity filters: id, creator, team
//   - 1 text filter: name
//   - 1 boolean filter: is-group
//
// All filters follow the per-filter file pattern and are registered
// in the Filters map for use with FromFlags.
package label
