package customview

const (
	mockCustomViewsResponse = `{
		"data": {
			"customViews": {
				"nodes": [
					{"id": "cv-123", "name": "My View", "shared": true, "modelName": "issue"}
				],
				"pageInfo": {"hasNextPage": false}
			}
		}
	}`

	mockCustomViewResponse = `{
		"data": {
			"customView": {
				"id": "cv-123",
				"name": "My View",
				"description": "A test view",
				"filterData": {"state": {"name": {"eq": "In Progress"}}},
				"shared": true,
				"modelName": "issue"
			}
		}
	}`

	mockCustomViewCreateResponse = `{
		"data": {
			"customViewCreate": {
				"success": true,
				"customView": {
					"id": "cv-999",
					"name": "New View",
					"shared": false
				}
			}
		}
	}`

	mockCustomViewUpdateResponse = `{
		"data": {
			"customViewUpdate": {
				"success": true,
				"customView": {
					"id": "cv-123",
					"name": "Updated View"
				}
			}
		}
	}`

	mockCustomViewDeleteResponse = `{
		"data": {
			"customViewDelete": {
				"success": true
			}
		}
	}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"ListCustomViews":  mockCustomViewsResponse,
		"GetCustomView":    mockCustomViewResponse,
		"CreateCustomView": mockCustomViewCreateResponse,
		"UpdateCustomView": mockCustomViewUpdateResponse,
		"DeleteCustomView": mockCustomViewDeleteResponse,
	}
}
