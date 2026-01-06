package user

const (
	mockUsersResponse = `{
		"data": {
			"users": {
				"nodes": [
					{"id": "user-123", "name": "Test User", "email": "test@example.com", "displayName": "Test", "createdAt": "2024-01-01T00:00:00.000Z", "active": true, "avatarUrl": "https://example.com/avatar.png"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockUserResponse = `{
		"data": {
			"user": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "Test",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-01T00:00:00.000Z",
				"admin": false,
				"active": true,
				"avatarUrl": "https://example.com/avatar.png"
			}
		}
	}`

	mockViewerResponse = `{
		"data": {
			"viewer": {
				"id": "viewer-123",
				"name": "Current User",
				"email": "current@example.com",
				"displayName": "Current",
				"active": true
			}
		}
	}`

	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [{"id": "team-123", "key": "ENG", "name": "Engineering"}],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssuesResponse = `{
		"data": {
			"issues": {
				"nodes": [
					{"id": "issue-123", "identifier": "ENG-123", "title": "Test Issue", "completedAt": "2024-01-01T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListUsers":  mockUsersResponse,
		"GetUser":    mockUserResponse,
		"GetViewer":  mockViewerResponse,
		"ListTeams":  mockTeamsResponse,
		"ListIssues": mockIssuesResponse,
	}
}
