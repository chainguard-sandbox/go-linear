package issue

const (
	mockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [
					{"id": "team-123", "key": "ENG", "name": "Engineering"}
				],
				"pageInfo": {"hasNextPage": false}
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

	mockStatesResponse = `{
		"data": {
			"workflowStates": {
				"nodes": [
					{"id": "state-123", "name": "Todo", "type": "unstarted"},
					{"id": "state-456", "name": "In Progress", "type": "started"},
					{"id": "state-789", "name": "Done", "type": "completed"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockLabelsResponse = `{
		"data": {
			"issueLabels": {
				"nodes": [
					{"id": "label-123", "name": "bug"},
					{"id": "label-456", "name": "feature"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueResponse = `{
		"data": {
			"issue": {
				"id": "issue-123",
				"identifier": "ENG-123",
				"title": "Test Issue",
				"description": "Test description",
				"priority": 2,
				"state": {"id": "state-456", "name": "In Progress"},
				"team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
				"assignee": {"id": "user-123", "name": "Test User"},
				"createdAt": "2024-01-01T00:00:00.000Z",
				"updatedAt": "2024-01-02T00:00:00.000Z"
			}
		}
	}`

	mockIssuesResponse = `{
		"data": {
			"issues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue",
						"priority": 2,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-01T00:00:00.000Z",
						"updatedAt": "2024-01-02T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueCreateResponse = `{
		"data": {
			"issueCreate": {
				"success": true,
				"issue": {
					"id": "issue-new",
					"identifier": "ENG-999",
					"title": "New Issue",
					"priority": 1,
					"number": 999,
					"url": "https://linear.app/test/issue/ENG-999",
					"state": {"id": "state-123", "name": "Todo"},
					"team": {"id": "team-123", "key": "ENG", "name": "Engineering"},
					"createdAt": "2024-01-01T00:00:00.000Z"
				}
			}
		}
	}`

	mockIssueUpdateResponse = `{
		"data": {
			"issueUpdate": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123",
					"title": "Updated Issue",
					"description": "Updated description",
					"priority": 1,
					"updatedAt": "2024-01-02T00:00:00.000Z"
				}
			}
		}
	}`

	mockIssueDeleteResponse = `{
		"data": {
			"issueDelete": {
				"success": true
			}
		}
	}`

	mockIssueArchiveResponse = `{
		"data": {
			"issueArchive": {
				"success": true
			}
		}
	}`

	mockIssueUnarchiveResponse = `{
		"data": {
			"issueUnarchive": {
				"success": true
			}
		}
	}`

	mockSearchResponse = `{
		"data": {
			"searchIssues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue with search term",
						"priority": 2,
						"state": {"id": "state-456", "name": "In Progress"},
						"team": {"id": "team-123", "key": "ENG"},
						"createdAt": "2024-01-01T00:00:00.000Z",
						"updatedAt": "2024-01-02T00:00:00.000Z"
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockIssueLabelResponse = `{
		"data": {
			"issueAddLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"title": "Test Issue",
					"labels": {
						"nodes": [{"id": "label-123", "name": "bug"}]
					}
				}
			}
		}
	}`

	mockIssueRemoveLabelResponse = `{
		"data": {
			"issueRemoveLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"title": "Test Issue",
					"labels": {"nodes": []}
				}
			}
		}
	}`

	mockIssueRelationCreateResponse = `{
		"data": {
			"issueRelationCreate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "blocks",
					"issue": {"id": "issue-123", "title": "Test Issue"},
					"relatedIssue": {"id": "issue-456", "title": "Related Issue"}
				}
			}
		}
	}`

	mockIssueRelationDeleteResponse = `{
		"data": {
			"issueRelationDelete": {
				"success": true
			}
		}
	}`

	mockIssueRelationUpdateResponse = `{
		"data": {
			"issueRelationUpdate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "related",
					"issue": {"id": "issue-123", "title": "Test Issue"},
					"relatedIssue": {"id": "issue-456", "title": "Related Issue"}
				}
			}
		}
	}`

	mockProjectsResponse = `{
		"data": {
			"projects": {
				"nodes": [
					{"id": "project-123", "name": "Test Project", "state": "started"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockCyclesResponse = `{
		"data": {
			"cycles": {
				"nodes": [
					{"id": "cycle-123", "name": "Sprint 1", "startsAt": "2024-01-01T00:00:00.000Z", "endsAt": "2024-01-14T00:00:00.000Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`
)

const mockTemplatesResponse = `{
	"data": {
		"templates": [
			{"id": "tpl-123", "name": "Bug Report", "templateData": "{\"priority\":1}", "type": "issue"}
		]
	}
}`

func defaultHandlers() map[string]string {
	return map[string]string{
		// Queries - match by query name in document
		"ListTeams":          mockTeamsResponse,
		"ListUsers":          mockUsersResponse,
		"Viewer":             mockViewerResponse,
		"ListWorkflowStates": mockStatesResponse,
		"ListLabels":         mockLabelsResponse,
		"GetIssue":           mockIssueResponse,
		"ListIssues":         mockIssuesResponse,
		"SearchIssues":       mockSearchResponse,
		"ListProjects":       mockProjectsResponse,
		"ListCycles":         mockCyclesResponse,
		"ListTemplates":      mockTemplatesResponse,

		// Mutations - match by operation name (gqlgenc uses these)
		"CreateIssue":         mockIssueCreateResponse,
		"UpdateIssue":         mockIssueUpdateResponse,
		"DeleteIssue":         mockIssueDeleteResponse,
		"ArchiveIssue":        mockIssueArchiveResponse,
		"UnarchiveIssue":      mockIssueUnarchiveResponse,
		"IssueAddLabel":       mockIssueLabelResponse,
		"IssueRemoveLabel":    mockIssueRemoveLabelResponse,
		"IssueRelationCreate": mockIssueRelationCreateResponse,
		"IssueRelationDelete": mockIssueRelationDeleteResponse,
		"IssueRelationUpdate": mockIssueRelationUpdateResponse,
	}
}
