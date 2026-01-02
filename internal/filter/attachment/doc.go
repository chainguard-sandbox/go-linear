// Package attachment provides filtering for Linear attachments.
//
// The package implements 9 filter types:
//   - 2 date filters: created-at, updated-at
//   - 3 entity filters: id, creator, source-type
//   - 4 text filters: title, subtitle, url
//
// All filters follow the per-filter file pattern and are registered
// in the Filters map for use with FromFlags.
package attachment
