package dashboard

type DashboardTrendPoint struct {
	Date                string `json:"date"`
	UserCount           int64  `json:"userCount"`
	EnrollJoinCount     int64  `json:"enrollJoinCount"`
	SurveyResponseCount int64  `json:"surveyResponseCount"`
	ExamRecordCount     int64  `json:"examRecordCount"`
}

type AdminHomeResponse struct {
	UserCnt      int64 `json:"userCnt"`
	EnrollCnt    int64 `json:"enrollCnt"`
	NewsCnt      int64 `json:"newsCnt"`
	JoinCnt      int64 `json:"joinCnt"`
	EventCnt     int64 `json:"eventCnt"`
	EventUserCnt int64 `json:"eventUserCnt"`
	MgrCnt       int64 `json:"mgrCnt"`

	SurveyCnt                     int64 `json:"surveyCnt"`
	TodaySurveyResponseCnt        int64 `json:"todaySurveyResponseCnt"`
	ExamCnt                       int64 `json:"examCnt"`
	TodayExamRecordCnt            int64 `json:"todayExamRecordCnt"`
	PendingUserCnt                int64 `json:"pendingUserCnt"`
	PendingWorkflowTaskCnt        int64 `json:"pendingWorkflowTaskCnt"`
	FailedWorkflowNotificationCnt int64 `json:"failedWorkflowNotificationCnt"`
	FailedScheduledRunCnt         int64 `json:"failedScheduledRunCnt"`
	PendingFeedbackCnt            int64 `json:"pendingFeedbackCnt"`
	UnboundDingTalkUserCnt        int64 `json:"unboundDingTalkUserCnt"`
	UnreadNotificationCnt         int64 `json:"unreadNotificationCnt"`

	TrendDays    int                   `json:"trendDays"`
	Trend        []DashboardTrendPoint `json:"trend"`
	UpdatedAt    int64                 `json:"updatedAt"`
	Capabilities map[string]bool       `json:"capabilities"`
}
