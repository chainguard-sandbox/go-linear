package favorite

const (
	mockFavoriteCreateResponse = `{"data": {"favoriteCreate": {"success": true, "favorite": {"id": "fav-123"}}}}`
	mockFavoriteDeleteResponse = `{"data": {"favoriteDelete": {"success": true}}}`
	mockIssuesResponse         = `{"data": {"issues": {"nodes": [{"id": "issue-123", "identifier": "ENG-123"}], "pageInfo": {"hasNextPage": false}}}}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"FavoriteCreate": mockFavoriteCreateResponse,
		"FavoriteDelete": mockFavoriteDeleteResponse,
		"ListIssues":     mockIssuesResponse,
	}
}
