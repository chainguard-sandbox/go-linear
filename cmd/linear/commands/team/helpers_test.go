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

	mockCyclesFilteredResponse = `{
		"data": {
			"cycles": {
				"nodes": [
					{
						"id": "cycle-1",
						"number": 1,
						"name": "Sprint 1",
						"startsAt": "2024-01-01T00:00:00.000Z",
						"endsAt": "2024-01-14T00:00:00.000Z",
						"completedAt": "2024-01-14T00:00:00.000Z",
						"progress": 0.95,
						"scopeHistory": [0, 10, 20, 30],
						"completedScopeHistory": [0, 5, 15, 28],
						"inProgressScopeHistory": [0, 3, 2, 1],
						"issueCountHistory": [0, 5, 10, 12],
						"completedIssueCountHistory": [0, 2, 8, 11],
						"team": {"id": "team-123", "key": "ENG", "name": "Engineering"}
					},
					{
						"id": "cycle-2",
						"number": 2,
						"name": "Sprint 2",
						"startsAt": "2024-01-15T00:00:00.000Z",
						"endsAt": "2024-01-28T00:00:00.000Z",
						"completedAt": "2024-01-28T00:00:00.000Z",
						"progress": 0.88,
						"scopeHistory": [0, 15, 25, 35],
						"completedScopeHistory": [0, 8, 18, 31],
						"inProgressScopeHistory": [0, 4, 3, 2],
						"issueCountHistory": [0, 8, 12, 14],
						"completedIssueCountHistory": [0, 4, 9, 12],
						"team": {"id": "team-123", "key": "ENG", "name": "Engineering"}
					},
					{
						"id": "cycle-3",
						"number": 3,
						"name": "Sprint 3",
						"startsAt": "2024-01-29T00:00:00.000Z",
						"endsAt": "2024-02-11T00:00:00.000Z",
						"completedAt": "2024-02-11T00:00:00.000Z",
						"progress": 1.0,
						"scopeHistory": [0, 12, 22, 25],
						"completedScopeHistory": [0, 6, 18, 25],
						"inProgressScopeHistory": [0, 5, 3, 0],
						"issueCountHistory": [0, 6, 10, 10],
						"completedIssueCountHistory": [0, 3, 8, 10],
						"team": {"id": "team-123", "key": "ENG", "name": "Engineering"}
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockTeamUnarchiveResponse = `{
		"data": {
			"teamUnarchive": {
				"success": true
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
	handlers["ListCyclesFiltered"] = mockCyclesFilteredResponse
	handlers["UnarchiveTeam"] = mockTeamUnarchiveResponse
	return handlers
}
