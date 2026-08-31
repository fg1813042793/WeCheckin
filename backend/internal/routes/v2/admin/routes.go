package admin

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
	admindept "wecheckin/backend/internal/handler/admin/department"
	admindict "wecheckin/backend/internal/handler/admin/dict"
	admindingtalk "wecheckin/backend/internal/handler/admin/dingtalk"
	adminenroll "wecheckin/backend/internal/handler/admin/enroll"
	adminevent "wecheckin/backend/internal/handler/admin/event"
	adminexam "wecheckin/backend/internal/handler/admin/exam"
	adminhome "wecheckin/backend/internal/handler/admin/home"
	adminmenu "wecheckin/backend/internal/handler/admin/menu"
	adminmgr "wecheckin/backend/internal/handler/admin/mgr"
	adminnews "wecheckin/backend/internal/handler/admin/news"
	adminpermission "wecheckin/backend/internal/handler/admin/permission"
	adminposition "wecheckin/backend/internal/handler/admin/position"
	adminrole "wecheckin/backend/internal/handler/admin/role"
	adminsetup "wecheckin/backend/internal/handler/admin/setup"
	adminsurvey "wecheckin/backend/internal/handler/admin/survey"
	adminuser "wecheckin/backend/internal/handler/admin/user"
	adminworkflow "wecheckin/backend/internal/handler/admin/workflow"
	adminmw "wecheckin/backend/internal/middleware/admin"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowinfra "wecheckin/backend/internal/modules/workflow/infrastructure"
	workflowhttp "wecheckin/backend/internal/modules/workflow/transport/httpadmin"
	"wecheckin/backend/internal/routes/v2/routeparam"
	"wecheckin/backend/pkg/database"
)

func Register(h *server.Hertz) {
	aMgr := adminmgr.NewAdminMgrHandler()
	h.POST("/api/v2/admin/auth/login", aMgr.AdminLogin)

	admin := h.Group("/api/v2/admin", adminmw.AdminAuth(), adminmw.AdminPerm())
	registerBaseRoutes(admin, aMgr)
	registerContentRoutes(admin)
	registerSystemRoutes(admin)
	registerSurveyRoutes(admin)
	registerExamRoutes(admin)
	registerWorkflowRoutes(admin)
}

func registerWorkflowRoutes(admin *route.RouterGroup) {
	aWorkflow := adminworkflow.NewAdminWorkflowHandler()
	db := database.GetDB()
	store := workflowinfra.NewGormStore(db)
	runtimeService := workflowapp.NewService(store, workflowinfra.NewAssigneeResolver(db), workflowinfra.NewRandomIDGenerator())
	runtimeHandler := workflowhttp.NewRuntimeHandler(runtimeService)

	admin.GET("/workflow-definitions", aWorkflow.List)
	admin.POST("/workflow-definitions", aWorkflow.Create)
	admin.GET("/workflow-definitions/:id", aWorkflow.Detail)
	admin.PUT("/workflow-definitions/:id", aWorkflow.Update)
	admin.DELETE("/workflow-definitions/:id", aWorkflow.Delete)
	admin.POST("/workflow-definitions/:id/validate", aWorkflow.Validate)
	admin.POST("/workflow-definitions/:id/publish", aWorkflow.Publish)
	admin.GET("/workflow-definitions/:id/versions", aWorkflow.Versions)
	admin.GET("/workflow-instances", runtimeHandler.ListInstances)
	admin.POST("/workflow-instances", runtimeHandler.StartInstance)
	admin.GET("/workflow-instances/:id", runtimeHandler.GetInstance)
	admin.POST("/workflow-instances/:id/cancel", runtimeHandler.CancelInstance)
	admin.GET("/workflow-tasks", runtimeHandler.ListTasks)
	admin.POST("/workflow-tasks/:id/complete", runtimeHandler.CompleteTask)
}

func registerBaseRoutes(admin *route.RouterGroup, aMgr *adminmgr.AdminMgrHandler) {
	aHome := adminhome.NewAdminHomeHandler()
	aSetup := adminsetup.NewAdminSetupHandler()
	aUser := adminuser.NewAdminUserHandler()
	aDingTalk := admindingtalk.NewAdminDingTalkHandler()

	admin.GET("/home", aHome.AdminHome)
	admin.DELETE("/home/recommendations", aHome.ClearVouchData)

	admin.GET("/managers", aMgr.GetMgrList)
	admin.POST("/managers", aMgr.InsertMgr)
	admin.GET("/managers/:id", routeparam.WithQueryID(aMgr.GetMgrDetail))
	admin.PUT("/managers/:id", routeparam.WithFormID(aMgr.EditMgr))
	admin.DELETE("/managers/:id", routeparam.WithFormID(aMgr.DelMgr))
	admin.DELETE("/managers", aMgr.DelMgrs)
	admin.PATCH("/managers/:id/status", routeparam.WithFormID(aMgr.StatusMgr))
	admin.PATCH("/managers/:id/password", routeparam.WithFormID(aMgr.PwdMgr))
	admin.GET("/admin-sessions", aMgr.GetOnlineAdmins)
	admin.POST("/admin-sessions/:id/force-offline", routeparam.WithFormID(aMgr.ForceOfflineAdmin))
	admin.POST("/admin-sessions/batch-force-offline", aMgr.BatchForceOfflineAdmin)
	admin.POST("/auth/logout", aMgr.AdminLogout)
	admin.PATCH("/me/password", aMgr.PwdMgr)

	admin.GET("/logs", aMgr.GetLogList)
	admin.DELETE("/logs", aMgr.ClearLog)

	admin.PUT("/settings", aSetup.SetSetup)
	admin.PUT("/settings/content", aSetup.SetContentSetup)
	admin.GET("/settings/mini-qr", aSetup.GenMiniQr)
	admin.GET("/settings/debug-token", aSetup.DebugTokenConfig)
	admin.GET("/dingtalk/settings", aDingTalk.GetSettings)
	admin.PUT("/dingtalk/settings", aDingTalk.SaveSettings)
	admin.POST("/dingtalk/settings/notification-test", aDingTalk.TestNotification)
	admin.GET("/dingtalk/user-bindings", aDingTalk.GetUserBindings)
	admin.POST("/dingtalk/user-bindings", aDingTalk.SaveUserBinding)
	admin.PATCH("/dingtalk/user-bindings/:id/status", routeparam.WithFormID(aDingTalk.StatusUserBinding))
	admin.DELETE("/dingtalk/user-bindings/:id", routeparam.WithFormID(aDingTalk.DeleteUserBinding))
	admin.GET("/dingtalk/perf-reviews", aDingTalk.GetPerfReviews)
	admin.DELETE("/dingtalk/perf-reviews", aDingTalk.DeletePerfReviews)
	admin.GET("/dingtalk/perf-reviews/:id", routeparam.WithQueryID(aDingTalk.GetPerfReviewDetail))
	admin.DELETE("/dingtalk/perf-reviews/:id", routeparam.WithFormID(aDingTalk.DeletePerfReview))
	admin.GET("/dingtalk/perf-histories", aDingTalk.GetPerfHistories)
	admin.DELETE("/dingtalk/perf-histories", aDingTalk.DeletePerfHistories)
	admin.DELETE("/dingtalk/perf-histories/:id", routeparam.WithFormID(aDingTalk.DeletePerfHistory))

	admin.GET("/users", aUser.GetUserList)
	admin.POST("/users", aUser.AddUser)
	admin.GET("/users/by-openid/:openid", routeparam.WithQueryParam("openid", "openid", aUser.GetUserDetail))
	admin.GET("/users/:id", routeparam.WithQueryID(aUser.GetUserByID))
	admin.PUT("/users/:id", routeparam.WithFormID(aUser.EditUser))
	admin.DELETE("/users/:id", routeparam.WithFormID(aUser.DelUser))
	admin.DELETE("/users", aUser.DelUsers)
	admin.PATCH("/users/:id/status", routeparam.WithFormID(aUser.StatusUser))
	admin.PATCH("/users/:id/password", routeparam.WithFormID(aUser.ResetPassword))
	admin.GET("/users/form-fields", aUser.GetUserFormFields)
	admin.PUT("/users/form-fields", aUser.SaveUserFormFields)
	admin.GET("/users/data", aUser.UserDataGet)
	admin.GET("/users/data/export", aUser.UserDataExport)
	admin.DELETE("/users/data/:id", routeparam.WithFormID(aUser.UserDataDel))
	admin.GET("/user-sessions", aUser.GetOnlineUsers)
	admin.POST("/user-sessions/:id/force-offline", routeparam.WithFormID(aUser.ForceOfflineUser))
	admin.POST("/user-sessions/batch-force-offline", aUser.BatchForceOfflineUser)
}

func registerContentRoutes(admin *route.RouterGroup) {
	aNews := adminnews.NewAdminNewsHandler()
	aEnroll := adminenroll.NewAdminEnrollHandler()
	aEvent := adminevent.NewAdminEventHandler()

	admin.GET("/news", aNews.GetAdminNewsList)
	admin.POST("/news", aNews.InsertNews)
	admin.GET("/news/:id", routeparam.WithQueryID(aNews.GetNewsDetail))
	admin.PUT("/news/:id", routeparam.WithFormID(aNews.EditNews))
	admin.DELETE("/news/:id", routeparam.WithFormID(aNews.DelNews))
	admin.DELETE("/news", aNews.DelNewsList)
	admin.PATCH("/news/:id/status", routeparam.WithFormID(aNews.StatusNews))
	admin.PATCH("/news/:id/recommendation", routeparam.WithFormID(aNews.VouchNews))
	admin.PATCH("/news/:id/sort", routeparam.WithFormID(aNews.SortNews))
	admin.PATCH("/news/:id/forms", routeparam.WithFormID(aNews.UpdateNewsForms))
	admin.PATCH("/news/:id/picture", routeparam.WithFormID(aNews.UpdateNewsPic))
	admin.PATCH("/news/:id/content", routeparam.WithFormID(aNews.UpdateNewsContent))

	admin.GET("/enrollments", aEnroll.GetAdminEnrollList)
	admin.POST("/enrollments", aEnroll.InsertEnroll)
	admin.GET("/enrollments/:id", routeparam.WithQueryID(aEnroll.GetEnrollDetail))
	admin.PUT("/enrollments/:id", routeparam.WithFormID(aEnroll.EditEnroll))
	admin.DELETE("/enrollments/:id", routeparam.WithFormID(aEnroll.DelEnroll))
	admin.DELETE("/enrollments", aEnroll.DelEnrolls)
	admin.PATCH("/enrollments/:id/status", routeparam.WithFormID(aEnroll.StatusEnroll))
	admin.PATCH("/enrollments/:id/sort", routeparam.WithFormID(aEnroll.SortEnroll))
	admin.PATCH("/enrollments/:id/recommendation", routeparam.WithFormID(aEnroll.VouchEnroll))
	admin.PATCH("/enrollments/:id/forms", routeparam.WithFormID(aEnroll.UpdateEnrollForms))
	admin.POST("/enrollments/:id/clear", routeparam.WithFormID(aEnroll.ClearEnrollAll))
	admin.GET("/enrollments/:id/joins", routeparam.WithQueryParam("enrollId", "id", aEnroll.GetEnrollJoinList))
	admin.DELETE("/enrollments/:id/joins/:joinId", routeparam.WithFormParams(map[string]string{"enrollJoinId": "joinId"}, aEnroll.DelEnrollJoin))
	admin.DELETE("/enrollments/:id/joins", aEnroll.DelEnrollJoins)
	admin.GET("/enrollments/:id/users", routeparam.WithQueryParam("enrollId", "id", aEnroll.GetEnrollUserList))
	admin.GET("/enrollments/:id/stats", routeparam.WithQueryParam("enrollId", "id", aEnroll.GetEnrollStats))
	admin.DELETE("/enrollments/:id/users/:userId", routeparam.WithFormParams(map[string]string{"enrollId": "id", "userId": "userId"}, aEnroll.RemoveEnrollUser))
	admin.DELETE("/enrollments/:id/users", routeparam.WithFormParam("enrollId", "id", aEnroll.RemoveEnrollUsers))
	admin.PUT("/enrollments/:id/users/:userId/forms", routeparam.WithFormParams(map[string]string{"enrollId": "id", "userId": "userId"}, aEnroll.EditEnrollUserForms))
	admin.GET("/enrollments/:id/export", routeparam.WithQueryParam("enrollId", "id", aEnroll.EnrollJoinDataGet))
	admin.POST("/enrollments/:id/export", routeparam.WithQueryParam("enrollId", "id", aEnroll.EnrollJoinDataExport))
	admin.DELETE("/enrollments/:id/export", routeparam.WithFormParam("enrollId", "id", aEnroll.EnrollJoinDataDel))

	admin.GET("/events", aEvent.GetAdminEventList)
	admin.POST("/events", aEvent.InsertEvent)
	admin.GET("/events/:id", routeparam.WithQueryID(aEvent.GetAdminEventDetail))
	admin.PUT("/events/:id", routeparam.WithFormID(aEvent.EditEvent))
	admin.DELETE("/events/:id", routeparam.WithFormID(aEvent.DelEvent))
	admin.DELETE("/events", aEvent.DelEvents)
	admin.PATCH("/events/:id/status", routeparam.WithFormID(aEvent.StatusEvent))
	admin.PATCH("/events/:id/recommendation", routeparam.WithFormID(aEvent.VouchEvent))
	admin.PATCH("/events/:id/top", routeparam.WithFormID(aEvent.TopEvent))
	admin.GET("/events/:id/participants", routeparam.WithQueryParam("eventId", "id", aEvent.GetEventParticipantList))
	admin.PUT("/events/:id/participants/:participantId", routeparam.WithFormParam("id", "participantId", aEvent.EditEventParticipant))
	admin.DELETE("/events/:id/participants/:participantId", routeparam.WithFormParam("id", "participantId", aEvent.DelEventParticipant))
	admin.DELETE("/events/:id/participants", aEvent.DelEventParticipants)
	admin.GET("/events/:id/dynamics", routeparam.WithQueryParam("eventId", "id", aEvent.GetEventDynamics))
	admin.POST("/events/:id/dynamics", routeparam.WithFormParam("eventId", "id", aEvent.PostEventDynamic))
	admin.PUT("/events/:id/dynamics/:dynamicId", routeparam.WithFormParam("id", "dynamicId", aEvent.EditEventDynamic))
	admin.DELETE("/events/:id/dynamics/:dynamicId", routeparam.WithFormParam("id", "dynamicId", aEvent.DelEventDynamic))
	admin.DELETE("/events/:id/dynamics", aEvent.DelEventDynamics)
	admin.GET("/events/:id/scores", routeparam.WithQueryParam("eventId", "id", aEvent.GetEventScores))
	admin.POST("/events/:id/scores", routeparam.WithFormParam("eventId", "id", aEvent.EditEventScore))
	admin.PUT("/events/:id/scores/:scoreId", routeparam.WithFormParam("id", "scoreId", aEvent.EditEventScore))
	admin.GET("/event-dept-users", aEvent.GetDeptUsers)
}

func registerSystemRoutes(admin *route.RouterGroup) {
	aDict := admindict.NewAdminDictHandler()
	aDept := admindept.NewAdminDeptHandler()
	aPosition := adminposition.NewAdminPositionHandler()
	aRole := adminrole.NewAdminRoleHandler()
	aMenu := adminmenu.NewAdminMenuHandler()
	aPermission := adminpermission.NewAdminPermissionHandler()

	admin.GET("/dict/types", aDict.GetDictTypes)
	admin.GET("/dict/items", aDict.GetDictByType)
	admin.POST("/dict/items", aDict.AddDictItem)
	admin.PUT("/dict/items/:id", routeparam.WithFormID(aDict.EditDictItem))
	admin.DELETE("/dict/items/:id", routeparam.WithFormID(aDict.DelDictItem))
	admin.DELETE("/dict/types/:typeCode/items", routeparam.WithFormParam("typeCode", "typeCode", aDict.DelDictByType))
	admin.PATCH("/dict/types/:typeCode", routeparam.WithFormParam("oldTypeCode", "typeCode", aDict.EditDictTypeName))

	admin.GET("/departments/tree", aDept.GetDeptTree)
	admin.POST("/departments", aDept.AddDept)
	admin.PUT("/departments/:id", routeparam.WithFormID(aDept.EditDept))
	admin.DELETE("/departments/:id", routeparam.WithFormID(aDept.DelDept))

	admin.GET("/positions", aPosition.GetPositionList)
	admin.POST("/positions", aPosition.AddPosition)
	admin.PUT("/positions/:id", routeparam.WithFormID(aPosition.EditPosition))
	admin.DELETE("/positions/:id", routeparam.WithFormID(aPosition.DelPosition))

	admin.GET("/roles/application-permissions", aRole.GetApplicationPermissionTree)
	admin.GET("/roles", aRole.GetRoleList)
	admin.POST("/roles", aRole.AddRole)
	admin.PUT("/roles/:id", routeparam.WithFormID(aRole.EditRole))
	admin.DELETE("/roles/:id", routeparam.WithFormID(aRole.DelRole))
	admin.DELETE("/roles", aRole.DelRoles)

	admin.GET("/me/menus", aMenu.GetAdminMenus)
	admin.GET("/me/perms", aMenu.GetAdminPerms)

	admin.GET("/permissions/tree", aPermission.GetPermissionTree)
	admin.GET("/permissions", aPermission.GetPermissionList)
	admin.POST("/permissions", aPermission.AddPermission)
	admin.PUT("/permissions/:key", aPermission.EditPermission)
	admin.DELETE("/permissions/:key", routeparam.WithFormParam("key", "key", aPermission.DelPermission))
}

func registerSurveyRoutes(admin *route.RouterGroup) {
	aSurvey := adminsurvey.NewAdminSurveyHandler()

	admin.GET("/survey-types", aSurvey.ListTypes)
	admin.POST("/survey-schema/parse", aSurvey.ParseSchema)
	admin.POST("/survey-expressions/evaluate", aSurvey.EvalExpr)
	admin.GET("/survey-report/enroll", aSurvey.ReportEnrollSchema)
	admin.GET("/survey-report/enroll/export", aSurvey.ExportEnrollSchemaCSV)
	admin.GET("/survey-report/event", aSurvey.ReportEventSchema)
	admin.GET("/survey-report/event/export", aSurvey.ExportEventSchemaCSV)
	admin.GET("/survey-report/survey", aSurvey.ReportSurveySchema)
	admin.GET("/survey-report/survey/export", aSurvey.ExportSurveySchemaCSV)

	admin.GET("/surveys", aSurvey.List)
	admin.POST("/surveys", aSurvey.Insert)
	admin.GET("/surveys/:id", routeparam.WithQueryID(aSurvey.Detail))
	admin.PUT("/surveys/:id", routeparam.WithBodyOrFormID(aSurvey.Edit))
	admin.DELETE("/surveys/:id", routeparam.WithFormID(aSurvey.Del))
	admin.PATCH("/surveys/:id/status", routeparam.WithFormID(aSurvey.Status))
	admin.POST("/surveys/:id/copy", routeparam.WithFormID(aSurvey.Copy))
	admin.GET("/surveys/:id/statistics", routeparam.WithQueryParam("surveyId", "id", aSurvey.Statistic))
	admin.GET("/surveys/:id/responses", routeparam.WithQueryParam("surveyId", "id", aSurvey.ResponseList))
	admin.GET("/surveys/:id/responses/export", routeparam.WithQueryParam("surveyId", "id", aSurvey.ResponseExport))
	admin.GET("/survey-responses/:id", routeparam.WithQueryID(aSurvey.ResponseDetail))
	admin.DELETE("/survey-responses/:id", routeparam.WithFormID(aSurvey.ResponseDel))
	admin.DELETE("/survey-responses", aSurvey.ResponseBatchDel)
	admin.GET("/surveys/:id/channels", routeparam.WithQueryParam("surveyId", "id", aSurvey.ChannelList))
	admin.POST("/surveys/:id/channels", routeparam.WithBodyOrFormParam("surveyId", "id", aSurvey.ChannelInsert))
	admin.DELETE("/survey-channels/:id", routeparam.WithFormID(aSurvey.ChannelDel))
	admin.POST("/survey-resources", aSurvey.ResourceUpload)
	admin.GET("/surveys/:id/resources", routeparam.WithQueryParam("surveyId", "id", aSurvey.ResourceList))
	admin.DELETE("/survey-resources/:id", routeparam.WithFormID(aSurvey.ResourceDelete))
	admin.GET("/survey-question-bank", aSurvey.QuestionBankList)
	admin.POST("/survey-question-bank", aSurvey.QuestionBankInsert)
	admin.PUT("/survey-question-bank/:id", routeparam.WithBodyOrFormID(aSurvey.QuestionBankEdit))
	admin.DELETE("/survey-question-bank/:id", routeparam.WithFormID(aSurvey.QuestionBankDel))
	admin.GET("/survey-question-bank/categories", aSurvey.QuestionBankCategories)
	admin.GET("/survey-notifications", aSurvey.NotifyList)
	admin.PATCH("/survey-notifications/:id/read", routeparam.WithFormID(aSurvey.NotifyRead))
	admin.GET("/survey-notifications/unread-count", aSurvey.NotifyUnreadCount)
	admin.GET("/survey-template-presets", aSurvey.TemplatePresetsGet)
	admin.PUT("/survey-template-presets", aSurvey.TemplatePresetsSave)
}

func registerExamRoutes(admin *route.RouterGroup) {
	aExam := adminexam.NewAdminExamHandler()

	admin.GET("/exams", aExam.List)
	admin.POST("/exams", aExam.Save)
	admin.GET("/exams/:id", routeparam.WithQueryID(aExam.Detail))
	admin.PUT("/exams/:id", routeparam.WithBodyOrFormID(aExam.Save))
	admin.DELETE("/exams/:id", routeparam.WithFormID(aExam.Delete))
	admin.PATCH("/exams/:id/status", routeparam.WithFormID(aExam.Status))
	admin.GET("/exams/:id/records", routeparam.WithQueryParam("examId", "id", aExam.RecordList))
	admin.GET("/exams/:id/records/:recordId", routeparam.WithQueryParam("id", "recordId", aExam.RecordDetail))
	admin.DELETE("/exams/:id/records/:recordId", routeparam.WithFormParam("id", "recordId", aExam.RecordDel))
	admin.DELETE("/exams/:id/records", aExam.RecordBatchDel)
	admin.GET("/exams/:id/statistics", routeparam.WithQueryParam("examId", "id", aExam.Statistics))
	admin.POST("/exam-resources", aExam.ResourceUpload)
	admin.GET("/exams/:id/resources", routeparam.WithQueryParam("examId", "id", aExam.ResourceList))
	admin.DELETE("/exam-resources/:id", routeparam.WithFormID(aExam.ResourceDelete))
	admin.GET("/exam-question-bank", aExam.QuestionBankList)
	admin.POST("/exam-question-bank", aExam.QuestionBankInsert)
	admin.PUT("/exam-question-bank/:id", routeparam.WithBodyOrFormID(aExam.QuestionBankEdit))
	admin.DELETE("/exam-question-bank/:id", routeparam.WithFormID(aExam.QuestionBankDel))
	admin.GET("/exam-question-bank/categories", aExam.QuestionBankCategories)
}
