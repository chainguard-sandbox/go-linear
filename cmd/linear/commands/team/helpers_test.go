package team

import "github.com/chainguard-sandbox/go-linear/internal/testutil"

const (
	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [
					{"id": "team-123", "key": "ENG", "name": "Engineering", "description": "Engineering team", "createdAt": "2024-01-01T00:00:00.000Z"},
					{"id": "team-456", "key": "PLT", "name": "Platform", "description": "Platform team", "createdAt": "2024-01-02T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false, "endCursor": null}
			}
		}
	}`

	mockTeamResponse = `{
		"data": {
			"team": {
				"id": "team-123",
				"key": "ENG",
				"name": "Engineering",
				"description": "Engineering team",
				"icon": "🛠",
				"createdAt": "2024-01-01T00:00:00.000Z",
				"color": "#3b82f6",
				"private": false
			}
		}
	}`

	mockTeamMembersResponse = `{
		"data": {
			"team": {
				"id": "team-123",
				"members": {
					"nodes": [
						{"id": "user-123", "name": "Test User", "email": "test@example.com"},
						{"id": "user-456", "name": "Another User", "email": "another@example.com"}
					],
					"pageInfo": {"hasNextPage": false}
				}
			}
		}
	}`

	mockTeamCreateResponse = `{
		"data": {
			"teamCreate": {
				"success": true,
				"team": {
					"id": "team-new",
					"key": "NEW",
					"name": "New Team",
					"description": "A new team",
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockTeamUpdateResponse = `{
		"data": {
			"teamUpdate": {
				"success": true,
				"team": {
					"id": "team-123",
					"key": "ENG",
					"name": "Updated Engineering",
					"description": "Updated description",
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockTeamDeleteResponse = `{
		"data": {
			"teamDelete": {
				"success": true
			}
		}
	}`

	mockUsersResponse = `{
		"data": {
			"users": {
				"nodes": [
					{"id": "user-123", "name": "Test User", "email": "test@example.com", "displayName": "test"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockViewerResponse = `{
		"data": {
			"viewer": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "test",
				"active": true,
				"admin": false,
				"createdAt": "2024-01-01T00:00:00.000Z"
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	handlers := testutil.DefaultHandlers()
	handlers["ListTeams"] = mockTeamsResponse
	handlers["GetTeam"] = mockTeamResponse
	handlers["ListUsers"] = mockUsersResponse
	handlers["Viewer"] = mockViewerResponse
	handlers["TeamMembers"] = mockTeamMembersResponse
	handlers["CreateTeam"] = mockTeamCreateResponse
	handlers["UpdateTeam"] = mockTeamUpdateResponse
	handlers["DeleteTeam"] = mockTeamDeleteResponse
	return handlers
}
