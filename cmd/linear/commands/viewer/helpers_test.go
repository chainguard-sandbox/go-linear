package viewer

const mockViewerResponse = `{
	"data": {
		"viewer": {
			"id": "viewer-123",
			"name": "Test User",
			"email": "test@example.com",
			"displayName": "Test",
			"createdAt": "2024-01-01T00:00:00.000Z",
			"admin": false,
			"active": true,
			"avatarUrl": "https://example.com/avatar.png"
		}
	}
}`

func defaultHandlers() map[string]string {
	return map[string]string{
		"Viewer": mockViewerResponse,
	}
}
