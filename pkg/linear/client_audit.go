package linear

import (
	"context"

	intgraphql "github.com/chainguard-sandbox/go-linear/v2/internal/graphql"
)

// AuditEntries retrieves a paginated list of audit log entries.
//
// Parameters:
//   - first: Number of entries to return (nil = server default ~50)
//   - after: Cursor for pagination (nil = start from beginning)
//   - filter: Optional filters (type, actor, ip, countryCode, createdAt)
//
// Returns:
//   - AuditEntries with nodes and pageInfo
//   - error: Non-nil if query fails
//
// Permissions Required: Admin
//
// Related: [AuditEntryTypes]
func (c *Client) AuditEntries(ctx context.Context, first *int64, after *string, filter *intgraphql.AuditEntryFilter) (*intgraphql.ListAuditEntries_AuditEntries, error) {
	resp, err := c.gqlClient.ListAuditEntries(ctx, first, after, filter)
	if err != nil {
		return nil, wrapGraphQLError("auditEntries query", err)
	}
	return &resp.AuditEntries, nil
}

// AuditEntryTypes retrieves the list of available audit entry types.
//
// Returns:
//   - Array of AuditEntryType with type and description
//   - error: Non-nil if query fails
//
// Permissions Required: Admin
//
// Related: [AuditEntries]
func (c *Client) AuditEntryTypes(ctx context.Context) ([]*intgraphql.ListAuditEntryTypes_AuditEntryTypes, error) {
	resp, err := c.gqlClient.ListAuditEntryTypes(ctx)
	if err != nil {
		return nil, wrapGraphQLError("auditEntryTypes query", err)
	}
	return resp.AuditEntryTypes, nil
}
