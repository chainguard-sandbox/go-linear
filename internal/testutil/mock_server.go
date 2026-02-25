// Package testutil provides shared test utilities for command tests.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chainguard-sandbox/go-linear/internal/cli"
	"github.com/chainguard-sandbox/go-linear/pkg/linear"
)

// MockServer creates a test server that handles Linear API queries.
// Handlers are matched by GraphQL operation name (case-insensitive).
func MockServer(t *testing.T, handlers map[string]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var reqBody struct {
			Query         string         `json:"query"`
			OperationName string         `json:"operationName"`
			Variables     map[string]any `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Logf("Failed to decode request: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Match by operation name (case-insensitive)
		for key, response := range handlers {
			if strings.EqualFold(key, reqBody.OperationName) {
				_, _ = w.Write([]byte(response))
				return
			}
		}

		// No match - log and return empty response
		t.Logf("No handler for operation: %s", reqBody.OperationName)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
}

// TestFactory creates a cli.ClientFactory that connects to the given test server.
func TestFactory(t *testing.T, serverURL string) cli.ClientFactory {
	t.Helper()
	return func() (*linear.Client, error) {
		return linear.NewClient("test_api_key", linear.WithBaseURL(serverURL))
	}
}

// Common mock responses for Linear API queries.
// Use GraphQL operation names as keys (e.g., "ListTeams", "GetIssue", "CreateIssue").
const (
	MockTeamsResponse = `{
		"data": {
			"teams": {
				"nodes": [
					{"id": "team-123", "key": "ENG", "name": "Engineering", "issueCount": 42}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockTeamResponse = `{
		"data": {
			"team": {
				"id": "team-123",
				"key": "ENG",
				"name": "Engineering",
				"issueCount": 42
			}
		}
	}`

	MockTeamMembersResponse = `{
		"data": {
			"team": {
				"members": {
					"nodes": [
						{"id": "user-123", "name": "Test User", "email": "test@example.com"}
					],
					"pageInfo": {"hasNextPage": false}
				}
			}
		}
	}`

	MockUsersResponse = `{
		"data": {
			"users": {
				"nodes": [
					{"id": "user-123", "name": "Test User", "email": "test@example.com", "displayName": "Test", "active": true}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockUserResponse = `{
		"data": {
			"user": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "Test",
				"active": true
			}
		}
	}`

	MockViewerResponse = `{
		"data": {
			"viewer": {
				"id": "user-123",
				"name": "Test User",
				"email": "test@example.com",
				"displayName": "Test",
				"active": true
			}
		}
	}`

	MockStatesResponse = `{
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

	MockLabelsResponse = `{
		"data": {
			"issueLabels": {
				"nodes": [
					{"id": "label-123", "name": "bug", "color": "#ff0000"},
					{"id": "label-456", "name": "feature", "color": "#00ff00"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockLabelResponse = `{
		"data": {
			"issueLabel": {
				"id": "label-123",
				"name": "bug",
				"color": "#ff0000"
			}
		}
	}`

	MockIssueResponse = `{
		"data": {
			"issue": {
				"id": "issue-123",
				"identifier": "ENG-123",
				"title": "Test Issue",
				"description": "Test description",
				"priority": 2,
				"state": {"id": "state-456", "name": "In Progress"},
				"team": {"id": "team-123", "key": "ENG"},
				"url": "https://linear.app/test/issue/ENG-123"
			}
		}
	}`

	MockIssuesResponse = `{
		"data": {
			"issues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue",
						"priority": 2,
						"state": {"name": "In Progress"},
						"team": {"key": "ENG"}
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockSearchResponse = `{
		"data": {
			"searchIssues": {
				"nodes": [
					{
						"id": "issue-123",
						"identifier": "ENG-123",
						"title": "Test Issue",
						"priority": 2,
						"state": {"name": "In Progress"},
						"team": {"key": "ENG"}
					}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockIssueCreateResponse = `{
		"data": {
			"issueCreate": {
				"success": true,
				"issue": {
					"id": "issue-999",
					"identifier": "ENG-999",
					"title": "New Issue",
					"url": "https://linear.app/test/issue/ENG-999"
				}
			}
		}
	}`

	MockIssueUpdateResponse = `{
		"data": {
			"issueUpdate": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123",
					"title": "Updated Issue"
				}
			}
		}
	}`

	MockIssueDeleteResponse = `{
		"data": {
			"issueDelete": {
				"success": true
			}
		}
	}`

	MockIssueLabelResponse = `{
		"data": {
			"issueAddLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123"
				}
			}
		}
	}`

	MockIssueRemoveLabelResponse = `{
		"data": {
			"issueRemoveLabel": {
				"success": true,
				"issue": {
					"id": "issue-123",
					"identifier": "ENG-123"
				}
			}
		}
	}`

	MockIssueRelationCreateResponse = `{
		"data": {
			"issueRelationCreate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "blocks"
				}
			}
		}
	}`

	MockIssueRelationDeleteResponse = `{
		"data": {
			"issueRelationDelete": {
				"success": true
			}
		}
	}`

	MockIssueRelationUpdateResponse = `{
		"data": {
			"issueRelationUpdate": {
				"success": true,
				"issueRelation": {
					"id": "relation-123",
					"type": "related"
				}
			}
		}
	}`

	MockCommentsResponse = `{
		"data": {
			"comments": {
				"nodes": [
					{"id": "comment-123", "body": "Test comment", "createdAt": "2024-01-01T00:00:00Z"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockCommentResponse = `{
		"data": {
			"comment": {
				"id": "comment-123",
				"body": "Test comment",
				"createdAt": "2024-01-01T00:00:00Z"
			}
		}
	}`

	MockCommentCreateResponse = `{
		"data": {
			"commentCreate": {
				"success": true,
				"comment": {
					"id": "comment-999",
					"body": "New comment"
				}
			}
		}
	}`

	MockCommentUpdateResponse = `{
		"data": {
			"commentUpdate": {
				"success": true,
				"comment": {
					"id": "comment-123",
					"body": "Updated comment"
				}
			}
		}
	}`

	MockCommentDeleteResponse = `{
		"data": {
			"commentDelete": {
				"success": true
			}
		}
	}`

	MockCyclesResponse = `{
		"data": {
			"cycles": {
				"nodes": [
					{"id": "cycle-123", "name": "Sprint 1", "number": 1}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockCycleResponse = `{
		"data": {
			"cycle": {
				"id": "cycle-123",
				"name": "Sprint 1",
				"number": 1
			}
		}
	}`

	MockCycleCreateResponse = `{
		"data": {
			"cycleCreate": {
				"success": true,
				"cycle": {
					"id": "cycle-999",
					"name": "New Sprint"
				}
			}
		}
	}`

	MockCycleUpdateResponse = `{
		"data": {
			"cycleUpdate": {
				"success": true,
				"cycle": {
					"id": "cycle-123",
					"name": "Updated Sprint"
				}
			}
		}
	}`

	MockCycleArchiveResponse = `{
		"data": {
			"cycleArchive": {
				"success": true
			}
		}
	}`

	MockProjectsResponse = `{
		"data": {
			"projects": {
				"nodes": [
					{"id": "project-123", "name": "Test Project", "state": "started"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockProjectResponse = `{
		"data": {
			"project": {
				"id": "project-123",
				"name": "Test Project",
				"state": "started"
			}
		}
	}`

	MockProjectCreateResponse = `{
		"data": {
			"projectCreate": {
				"success": true,
				"project": {
					"id": "project-999",
					"name": "New Project"
				}
			}
		}
	}`

	MockProjectUpdateResponse = `{
		"data": {
			"projectUpdate": {
				"success": true,
				"project": {
					"id": "project-123",
					"name": "Updated Project"
				}
			}
		}
	}`

	MockProjectDeleteResponse = `{
		"data": {
			"projectDelete": {
				"success": true
			}
		}
	}`

	MockProjectMilestoneCreateResponse = `{
		"data": {
			"projectMilestoneCreate": {
				"success": true,
				"projectMilestone": {
					"id": "milestone-999",
					"name": "New Milestone"
				}
			}
		}
	}`

	MockProjectMilestoneUpdateResponse = `{
		"data": {
			"projectMilestoneUpdate": {
				"success": true,
				"projectMilestone": {
					"id": "milestone-123",
					"name": "Updated Milestone"
				}
			}
		}
	}`

	MockProjectMilestoneDeleteResponse = `{
		"data": {
			"projectMilestoneDelete": {
				"success": true
			}
		}
	}`

	MockAttachmentsResponse = `{
		"data": {
			"attachments": {
				"nodes": [
					{"id": "attachment-123", "title": "Test Attachment", "url": "https://example.com"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockAttachmentResponse = `{
		"data": {
			"attachment": {
				"id": "attachment-123",
				"title": "Test Attachment",
				"url": "https://example.com"
			}
		}
	}`

	MockAttachmentCreateResponse = `{
		"data": {
			"attachmentCreate": {
				"success": true,
				"attachment": {
					"id": "attachment-999",
					"title": "New Attachment"
				}
			}
		}
	}`

	MockAttachmentDeleteResponse = `{
		"data": {
			"attachmentDelete": {
				"success": true
			}
		}
	}`

	MockAttachmentLinkURLResponse = `{
		"data": {
			"attachmentLinkURL": {
				"success": true,
				"attachment": {
					"id": "attachment-999",
					"url": "https://example.com"
				}
			}
		}
	}`

	MockAttachmentLinkGitHubPRResponse = `{
		"data": {
			"attachmentLinkGitHubPR": {
				"success": true,
				"attachment": {
					"id": "attachment-999"
				}
			}
		}
	}`

	MockAttachmentLinkSlackResponse = `{
		"data": {
			"attachmentLinkSlack": {
				"success": true,
				"attachment": {
					"id": "attachment-999"
				}
			}
		}
	}`

	MockDocumentsResponse = `{
		"data": {
			"documents": {
				"nodes": [
					{"id": "doc-123", "title": "Test Document", "content": "Test content"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockDocumentResponse = `{
		"data": {
			"document": {
				"id": "doc-123",
				"title": "Test Document",
				"content": "Test content"
			}
		}
	}`

	MockInitiativesResponse = `{
		"data": {
			"initiatives": {
				"nodes": [
					{"id": "initiative-123", "name": "Test Initiative"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockInitiativeResponse = `{
		"data": {
			"initiative": {
				"id": "initiative-123",
				"name": "Test Initiative"
			}
		}
	}`

	MockInitiativeCreateResponse = `{
		"data": {
			"initiativeCreate": {
				"success": true,
				"initiative": {
					"id": "initiative-999",
					"name": "New Initiative"
				}
			}
		}
	}`

	MockInitiativeUpdateResponse = `{
		"data": {
			"initiativeUpdate": {
				"success": true,
				"initiative": {
					"id": "initiative-123",
					"name": "Updated Initiative"
				}
			}
		}
	}`

	MockRoadmapsResponse = `{
		"data": {
			"roadmaps": {
				"nodes": [
					{"id": "roadmap-123", "name": "Test Roadmap"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	MockRoadmapResponse = `{
		"data": {
			"roadmap": {
				"id": "roadmap-123",
				"name": "Test Roadmap"
			}
		}
	}`

	MockTemplatesResponse = `{
		"data": {
			"templates": [
				{"id": "template-123", "name": "Bug Report", "templateData": "{\"priority\":1}", "type": "issue"}
			]
		}
	}`

	MockTemplateResponse = `{
		"data": {
			"template": {
				"id": "template-123",
				"name": "Bug Report",
				"templateData": "{\"priority\":1}",
				"type": "issue"
			}
		}
	}`

	MockOrganizationResponse = `{
		"data": {
			"organization": {
				"id": "org-123",
				"name": "Test Org",
				"urlKey": "test"
			}
		}
	}`

	MockFavoriteCreateResponse = `{
		"data": {
			"favoriteCreate": {
				"success": true,
				"favorite": {
					"id": "favorite-999"
				}
			}
		}
	}`

	MockFavoriteDeleteResponse = `{
		"data": {
			"favoriteDelete": {
				"success": true
			}
		}
	}`

	MockReactionCreateResponse = `{
		"data": {
			"reactionCreate": {
				"success": true,
				"reaction": {
					"id": "reaction-999",
					"emoji": "👍"
				}
			}
		}
	}`

	MockReactionDeleteResponse = `{
		"data": {
			"reactionDelete": {
				"success": true
			}
		}
	}`

	MockNotificationArchiveResponse = `{
		"data": {
			"notificationArchive": {
				"success": true
			}
		}
	}`

	MockNotificationUpdateResponse = `{
		"data": {
			"notificationUpdate": {
				"success": true,
				"notification": {
					"id": "notification-123"
				}
			}
		}
	}`

	MockNotificationSubscriptionCreateResponse = `{
		"data": {
			"notificationSubscriptionCreate": {
				"success": true,
				"notificationSubscription": {
					"id": "subscription-999"
				}
			}
		}
	}`

	MockNotificationSubscriptionDeleteResponse = `{
		"data": {
			"notificationSubscriptionDelete": {
				"success": true
			}
		}
	}`

	MockLabelCreateResponse = `{
		"data": {
			"issueLabelCreate": {
				"success": true,
				"issueLabel": {
					"id": "label-999",
					"name": "new-label",
					"color": "#0000ff"
				}
			}
		}
	}`

	MockLabelUpdateResponse = `{
		"data": {
			"issueLabelUpdate": {
				"success": true,
				"issueLabel": {
					"id": "label-123",
					"name": "updated-label"
				}
			}
		}
	}`

	MockLabelDeleteResponse = `{
		"data": {
			"issueLabelDelete": {
				"success": true
			}
		}
	}`

	MockTeamCreateResponse = `{
		"data": {
			"teamCreate": {
				"success": true,
				"team": {
					"id": "team-999",
					"key": "NEW",
					"name": "New Team"
				}
			}
		}
	}`

	MockTeamUpdateResponse = `{
		"data": {
			"teamUpdate": {
				"success": true,
				"team": {
					"id": "team-123",
					"name": "Updated Team"
				}
			}
		}
	}`

	MockTeamDeleteResponse = `{
		"data": {
			"teamDelete": {
				"success": true
			}
		}
	}`

	MockIssueBatchUpdateResponse = `{
		"data": {
			"issueBatchUpdate": {
				"success": true,
				"issues": [
					{"id": "issue-123"},
					{"id": "issue-456"}
				]
			}
		}
	}`

	MockStatusResponse = `{
		"data": {
			"viewer": {
				"id": "user-123"
			}
		}
	}`
)

// DefaultHandlers returns a map of operation names to mock responses.
// All keys are GraphQL operation names (e.g., "ListTeams", "CreateIssue").
func DefaultHandlers() map[string]string {
	return map[string]string{
		// Queries
		"ListTeams":          MockTeamsResponse,
		"GetTeam":            MockTeamResponse,
		"TeamMembers":        MockTeamMembersResponse,
		"ListUsers":          MockUsersResponse,
		"GetUser":            MockUserResponse,
		"Viewer":             MockViewerResponse,
		"ListWorkflowStates": MockStatesResponse,
		"GetWorkflowState":   MockStatesResponse,
		"ListLabels":         MockLabelsResponse,
		"GetLabel":           MockLabelResponse,
		"GetIssue":           MockIssueResponse,
		"ListIssues":         MockIssuesResponse,
		"SearchIssues":       MockSearchResponse,
		"ListComments":       MockCommentsResponse,
		"GetComment":         MockCommentResponse,
		"ListCycles":         MockCyclesResponse,
		"GetCycle":           MockCycleResponse,
		"ListProjects":       MockProjectsResponse,
		"GetProject":         MockProjectResponse,
		"ListAttachments":    MockAttachmentsResponse,
		"GetAttachment":      MockAttachmentResponse,
		"ListDocuments":      MockDocumentsResponse,
		"GetDocument":        MockDocumentResponse,
		"ListInitiatives":    MockInitiativesResponse,
		"GetInitiative":      MockInitiativeResponse,
		"ListRoadmaps":       MockRoadmapsResponse,
		"GetRoadmap":         MockRoadmapResponse,
		"ListTemplates":      MockTemplatesResponse,
		"GetTemplate":        MockTemplateResponse,
		"GetOrganization":    MockOrganizationResponse,

		// Issue mutations
		"CreateIssue":         MockIssueCreateResponse,
		"UpdateIssue":         MockIssueUpdateResponse,
		"DeleteIssue":         MockIssueDeleteResponse,
		"IssueAddLabel":       MockIssueLabelResponse,
		"IssueRemoveLabel":    MockIssueRemoveLabelResponse,
		"IssueRelationCreate": MockIssueRelationCreateResponse,
		"IssueRelationDelete": MockIssueRelationDeleteResponse,
		"IssueRelationUpdate": MockIssueRelationUpdateResponse,
		"IssueBatchUpdate":    MockIssueBatchUpdateResponse,

		// Comment mutations
		"CommentCreate": MockCommentCreateResponse,
		"CommentUpdate": MockCommentUpdateResponse,
		"CommentDelete": MockCommentDeleteResponse,

		// Cycle mutations
		"CycleCreate":  MockCycleCreateResponse,
		"CycleUpdate":  MockCycleUpdateResponse,
		"CycleArchive": MockCycleArchiveResponse,

		// Project mutations
		"ProjectCreate":          MockProjectCreateResponse,
		"ProjectUpdate":          MockProjectUpdateResponse,
		"ProjectDelete":          MockProjectDeleteResponse,
		"ProjectMilestoneCreate": MockProjectMilestoneCreateResponse,
		"ProjectMilestoneUpdate": MockProjectMilestoneUpdateResponse,
		"ProjectMilestoneDelete": MockProjectMilestoneDeleteResponse,

		// Attachment mutations
		"AttachmentCreate":       MockAttachmentCreateResponse,
		"AttachmentDelete":       MockAttachmentDeleteResponse,
		"AttachmentLinkURL":      MockAttachmentLinkURLResponse,
		"AttachmentLinkGitHubPR": MockAttachmentLinkGitHubPRResponse,
		"AttachmentLinkSlack":    MockAttachmentLinkSlackResponse,

		// Initiative mutations
		"InitiativeCreate": MockInitiativeCreateResponse,
		"InitiativeUpdate": MockInitiativeUpdateResponse,

		// Favorite mutations
		"FavoriteCreate": MockFavoriteCreateResponse,
		"FavoriteDelete": MockFavoriteDeleteResponse,

		// Reaction mutations
		"ReactionCreate": MockReactionCreateResponse,
		"ReactionDelete": MockReactionDeleteResponse,

		// Notification mutations
		"NotificationArchive":            MockNotificationArchiveResponse,
		"NotificationUpdate":             MockNotificationUpdateResponse,
		"NotificationSubscriptionCreate": MockNotificationSubscriptionCreateResponse,
		"NotificationSubscriptionDelete": MockNotificationSubscriptionDeleteResponse,

		// Label mutations
		"IssueLabelCreate": MockLabelCreateResponse,
		"IssueLabelUpdate": MockLabelUpdateResponse,
		"IssueLabelDelete": MockLabelDeleteResponse,

		// Team mutations
		"TeamCreate": MockTeamCreateResponse,
		"TeamUpdate": MockTeamUpdateResponse,
		"TeamDelete": MockTeamDeleteResponse,
	}
}
