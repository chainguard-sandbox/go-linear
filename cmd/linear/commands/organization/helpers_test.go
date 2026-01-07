package organization

const mockOrganizationResponse = `{
	"data": {
		"organization": {
			"id": "org-123",
			"name": "Test Org",
			"urlKey": "test-org",
			"logoUrl": "https://example.com/logo.png",
			"createdAt": "2024-01-01T00:00:00.000Z",
			"updatedAt": "2024-01-01T00:00:00.000Z",
			"roadmapEnabled": true
		}
	}
}`

func defaultHandlers() map[string]string {
	return map[string]string{
		"GetOrganization": mockOrganizationResponse,
	}
}
