package notification

const (
	mockNotificationArchiveResponse       = `{"data": {"notificationArchive": {"success": true}}}`
	mockNotificationUpdateResponse        = `{"data": {"notificationUpdate": {"success": true, "notification": {"id": "notif-123", "readAt": "2024-01-01T00:00:00.000Z"}}}}`
	mockSubscribeResponse                 = `{"data": {"notificationSubscriptionCreate": {"success": true, "notificationSubscription": {"id": "sub-123"}}}}`
	mockUnsubscribeResponse               = `{"data": {"notificationSubscriptionDelete": {"success": true}}}`
	mockProjectsResponse                  = `{"data": {"projects": {"nodes": [{"id": "proj-123", "name": "Test Project"}], "pageInfo": {"hasNextPage": false}}}}`
	mockNotificationArchiveAllResponse    = `{"data": {"notificationArchiveAll": {"success": true}}}`
	mockNotificationMarkReadAllResponse   = `{"data": {"notificationMarkReadAll": {"success": true}}}`
	mockNotificationMarkUnreadAllResponse = `{"data": {"notificationMarkUnreadAll": {"success": true}}}`
	mockNotificationSnoozeAllResponse     = `{"data": {"notificationSnoozeAll": {"success": true}}}`
	mockNotificationUnsnoozeAllResponse   = `{"data": {"notificationUnsnoozeAll": {"success": true}}}`
	mockIssuesResponse                    = `{"data": {"issues": {"nodes": [{"id": "issue-123", "identifier": "ENG-123", "title": "Test"}], "pageInfo": {"hasNextPage": false}}}}`
)

func defaultHandlers() map[string]string {
	return map[string]string{
		"NotificationArchive":            mockNotificationArchiveResponse,
		"NotificationUpdate":             mockNotificationUpdateResponse,
		"NotificationSubscriptionCreate": mockSubscribeResponse,
		"NotificationSubscriptionDelete": mockUnsubscribeResponse,
		"ListProjects":                   mockProjectsResponse,
		"NotificationArchiveAll":         mockNotificationArchiveAllResponse,
		"NotificationMarkReadAll":        mockNotificationMarkReadAllResponse,
		"NotificationMarkUnreadAll":      mockNotificationMarkUnreadAllResponse,
		"NotificationSnoozeAll":          mockNotificationSnoozeAllResponse,
		"NotificationUnsnoozeAll":        mockNotificationUnsnoozeAllResponse,
		"SearchIssues":                   mockIssuesResponse,
		"ListIssues":                     mockIssuesResponse,
	}
}
