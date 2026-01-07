package reaction

const (
	mockReactionCreateResponse = `{"data": {"reactionCreate": {"success": true, "reaction": {"id": "react-123", "emoji": "👍"}}}}`
	mockReactionDeleteResponse = `{"data": {"reactionDelete": {"success": true}}}`
	mockIssuesResponse         = `{"data": {"issues": {"nodes": [{"id": "issue-123", "identifier": "ENG-123"}], "pageInfo": {"hasNextPage": false}}}}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ReactionCreate": mockReactionCreateResponse,
		"ReactionDelete": mockReactionDeleteResponse,
		"ListIssues":     mockIssuesResponse,
	}
}
