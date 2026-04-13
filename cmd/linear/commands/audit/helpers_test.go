package audit

const (
	mockAuditEntriesResponse = `{
		"data": {
			"auditEntries": {
				"nodes": [
					{
						"id": "ae-123",
						"type": "issue.create",
						"createdAt": "2024-01-01T00:00:00.000Z",
						"ip": "192.168.1.1",
						"actor": {"id": "user-123", "name": "Test User"}
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockAuditEntryTypesResponse = `{
		"data": {
			"auditEntryTypes": [
				{"type": "issue.create", "description": "Issue created"},
				{"type": "issue.update", "description": "Issue updated"},
				{"type": "issue.delete", "description": "Issue deleted"}
			]
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListAuditEntries":    mockAuditEntriesResponse,
		"ListAuditEntryTypes": mockAuditEntryTypesResponse,
	}
}
