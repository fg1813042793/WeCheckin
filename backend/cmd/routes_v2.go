package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/route"
	admindept "wecheckin-backend/backend/internal/app/handler/admin/department"
	admindict "wecheckin-backend/backend/internal/app/handler/admin/dict"
	adminenroll "wecheckin-backend/backend/internal/app/handler/admin/enroll"
	adminevent "wecheckin-backend/backend/internal/app/handler/admin/event"
	adminexam "wecheckin-backend/backend/internal/app/handler/admin/exam"
	adminhome "wecheckin-backend/backend/internal/app/handler/admin/home"
	adminmenu "wecheckin-backend/backend/internal/app/handler/admin/menu"
	adminmgr "wecheckin-backend/backend/internal/app/handler/admin/mgr"
	adminnews "wecheckin-backend/backend/internal/app/handler/admin/news"
	adminpermission "wecheckin-backend/backend/internal/app/handler/admin/permission"
	adminrole "wecheckin-backend/backend/internal/app/handler/admin/role"
	adminsetup "wecheckin-backend/backend/internal/app/handler/admin/setup"
	adminsurvey "wecheckin-backend/backend/internal/app/handler/admin/survey"
	adminuser "wecheckin-backend/backend/internal/app/handler/admin/user"
	clientenroll "wecheckin-backend/backend/internal/app/handler/client/enroll"
	clientevent "wecheckin-backend/backend/internal/app/handler/client/event"
	clientexam "wecheckin-backend/backend/internal/app/handler/client/exam"
	clientfavorite "wecheckin-backend/backend/internal/app/handler/client/favorite"
	clientnews "wecheckin-backend/backend/internal/app/handler/client/news"
	clientpassport "wecheckin-backend/backend/internal/app/handler/client/passport"
	clientsurvey "wecheckin-backend/backend/internal/app/handler/client/survey"
	publicgeo "wecheckin-backend/backend/internal/app/handler/public/geo"
	publichome "wecheckin-backend/backend/internal/app/handler/public/home"
	"wecheckin-backend/backend/internal/middleware"
)

func registerV2Routes(h *server.Hertz) {
	registerV2PublicRoutes(h)
	registerV2ClientRoutes(h)
	registerV2AdminRoutes(h)
	registerV2DingTalkH5Routes(h)
}

func registerV2PublicRoutes(h *server.Hertz) {
	hm := publichome.NewHomeHandler()
	pp := clientpassport.NewPassportHandler()
	geo := publicgeo.NewGeoHandler()
	aUser := adminuser.NewAdminUserHandler()
	aDict := admindict.NewAdminDictHandler()
	ev := clientevent.NewEventHandler()
	cSurvey := clientsurvey.NewClientSurveyHandler()
	cExam := clientexam.NewClientExamHandler()

	h.GET("/api/v2/home", hm.GetHomeList)
	h.GET("/api/v2/home/setup", hm.GetSetup)
	h.GET("/api/v2/user-form-fields", aUser.GetUserFormFields)
	h.POST("/api/v2/auth/login", pp.Login)
	h.POST("/api/v2/auth/password-login", pp.LoginByPwd)
	h.POST("/api/v2/auth/register", pp.Register)
	h.GET("/api/v2/geo/reverse", geo.ReverseGeocode)
	h.GET("/api/v2/dict/types", aDict.GetDictTypes)
	h.GET("/api/v2/dict/items", aDict.GetDictByType)
	h.GET("/api/v2/events", ev.GetEventList)
	h.GET("/api/v2/events/:id", withQueryID(ev.ViewEvent))
	h.GET("/api/v2/surveys", cSurvey.List)
	h.GET("/api/v2/surveys/:id", withQueryID(cSurvey.Detail))
	h.POST("/api/v2/surveys/:id/responses", withBodyOrFormParam("surveyId", "id", cSurvey.Submit))
	h.POST("/api/v2/survey/apply", cSurvey.PublicApply)
	h.POST("/api/v2/survey/validate", cSurvey.PublicValidate)
	h.GET("/api/v2/exams", cExam.List)
	h.GET("/api/v2/exams/:id", withQueryID(cExam.View))
	h.POST("/api/v2/exams/:id/submissions", withBodyOrFormParam("examId", "id", cExam.Submit))
	h.POST("/api/v2/exams/:id/validation", withBodyOrFormParam("examId", "id", cExam.Validate))
	h.GET("/api/v2/exam-results", cExam.ResultBySession)
}

func registerV2ClientRoutes(h *server.Hertz) {
	pp := clientpassport.NewPassportHandler()
	ns := clientnews.NewNewsHandler()
	el := clientenroll.NewEnrollHandler()
	fa := clientfavorite.NewFavHandler()
	ev := clientevent.NewEventHandler()
	cSurvey := clientsurvey.NewClientSurveyHandler()
	cExam := clientexam.NewClientExamHandler()

	client := h.Group("/api/v2", middleware.ClientAuth(), middleware.ClientPerm())
	client.GET("/me", pp.GetMyDetail)
	client.PUT("/me", pp.EditBase)
	client.POST("/me/phone", pp.GetPhone)
	client.POST("/me/logout", pp.Logout)
	client.GET("/me/favorites", fa.GetMyFavList)
	client.POST("/me/favorites", fa.UpdateFav)
	client.DELETE("/me/favorites/:oid", withFormParam("oid", "oid", fa.DelFav))
	client.GET("/me/favorites/check", fa.IsFav)
	client.GET("/me/enrollments", el.GetMyEnrollJoinList)
	client.GET("/me/enrollment-users", el.GetMyEnrollUserList)
	client.GET("/me/enrollment-records", el.GetMyJoinRecords)
	client.GET("/me/enrollment-calendar", el.GetMyCalendar)
	client.GET("/me/enrollment-day-records", el.GetMyDayRecords)
	client.GET("/me/events", ev.GetMyEventList)
	client.GET("/me/event-roles", ev.GetMyEventRoles)
	client.GET("/me/managed-events", ev.GetMyManagedList)
	client.GET("/me/survey-responses", cSurvey.MyResponses)
	client.GET("/me/survey-responses/:id", withQueryID(cSurvey.MyResponseDetail))
	client.GET("/me/exam-records", cExam.MyRecords)

	client.GET("/news", ns.GetNewsList)
	client.GET("/news/categories", ns.GetNewsCateList)
	client.GET("/news/:id", withQueryID(ns.ViewNews))
	client.GET("/enrollments", el.GetEnrollList)
	client.GET("/enrollments/:id", withQueryID(el.ViewEnroll))
	client.GET("/enrollments/:id/join-days", withQueryID(el.GetEnrollJoinByDay))
	client.POST("/enrollments/:id/joins", withFormParam("enroll_id", "id", el.EnrollJoin))
	client.POST("/enrollments/:id/submissions", withFormParam("enroll_id", "id", el.EnrollUserSubmit))
	client.POST("/events/:id/participants", withFormParam("event_id", "id", ev.EventParticipate))
	client.GET("/events/:id/participants", withQueryParam("event_id", "id", ev.GetEventParticipantList))
	client.GET("/events/:id/dynamics", withQueryParam("event_id", "id", ev.GetEventDynamics))
	client.POST("/events/:id/dynamics", withFormParam("event_id", "id", ev.PostEventDynamic))
	client.GET("/events/:id/scores", withQueryParam("event_id", "id", ev.GetEventScores))
	client.POST("/events/:id/scores", withFormParam("event_id", "id", ev.SaveEventScore))
	client.POST("/exams/:id/start", withQueryParam("examId", "id", cExam.Start))
	client.GET("/exam-records/:id", withQueryID(cExam.Record))
	client.PUT("/exam-records/:id/answers", withFormParam("recordId", "id", cExam.SaveAnswer))
}

func registerV2AdminRoutes(h *server.Hertz) {
	aMgr := adminmgr.NewAdminMgrHandler()
	h.POST("/api/v2/admin/auth/login", aMgr.AdminLogin)

	admin := h.Group("/api/v2/admin", middleware.AdminAuth(), middleware.AdminPerm())
	registerV2AdminBaseRoutes(admin, aMgr)
	registerV2AdminContentRoutes(admin)
	registerV2AdminSystemRoutes(admin)
	registerV2AdminSurveyRoutes(admin)
	registerV2AdminExamRoutes(admin)
}

func registerV2AdminBaseRoutes(admin *route.RouterGroup, aMgr *adminmgr.AdminMgrHandler) {
	aHome := adminhome.NewAdminHomeHandler()
	aSetup := adminsetup.NewAdminSetupHandler()
	aUser := adminuser.NewAdminUserHandler()

	admin.GET("/home", aHome.AdminHome)
	admin.DELETE("/home/recommendations", aHome.ClearVouchData)

	admin.GET("/managers", aMgr.GetMgrList)
	admin.POST("/managers", aMgr.InsertMgr)
	admin.GET("/managers/:id", withQueryID(aMgr.GetMgrDetail))
	admin.PUT("/managers/:id", withFormID(aMgr.EditMgr))
	admin.DELETE("/managers/:id", withFormID(aMgr.DelMgr))
	admin.DELETE("/managers", aMgr.DelMgrs)
	admin.PATCH("/managers/:id/status", withFormID(aMgr.StatusMgr))
	admin.PATCH("/managers/:id/password", withFormID(aMgr.PwdMgr))
	admin.GET("/admin-sessions", aMgr.GetOnlineAdmins)
	admin.POST("/admin-sessions/:id/force-offline", withFormID(aMgr.ForceOfflineAdmin))
	admin.POST("/admin-sessions/batch-force-offline", aMgr.BatchForceOfflineAdmin)
	admin.POST("/auth/logout", aMgr.AdminLogout)
	admin.PATCH("/me/password", aMgr.PwdMgr)

	admin.GET("/logs", aMgr.GetLogList)
	admin.DELETE("/logs", aMgr.ClearLog)

	admin.PUT("/settings", aSetup.SetSetup)
	admin.PUT("/settings/content", aSetup.SetContentSetup)
	admin.GET("/settings/mini-qr", aSetup.GenMiniQr)
	admin.GET("/settings/debug-token", aSetup.DebugTokenConfig)

	admin.GET("/users", aUser.GetUserList)
	admin.POST("/users", aUser.AddUser)
	admin.GET("/users/by-openid/:openid", withQueryParam("openid", "openid", aUser.GetUserDetail))
	admin.GET("/users/:id", withQueryID(aUser.GetUserByID))
	admin.PUT("/users/:id", withFormID(aUser.EditUser))
	admin.DELETE("/users/:id", withFormID(aUser.DelUser))
	admin.DELETE("/users", aUser.DelUsers)
	admin.PATCH("/users/:id/status", withFormID(aUser.StatusUser))
	admin.PATCH("/users/:id/password", withFormID(aUser.ResetPassword))
	admin.GET("/users/form-fields", aUser.GetUserFormFields)
	admin.PUT("/users/form-fields", aUser.SaveUserFormFields)
	admin.GET("/users/data", aUser.UserDataGet)
	admin.GET("/users/data/export", aUser.UserDataExport)
	admin.DELETE("/users/data/:id", withFormID(aUser.UserDataDel))
	admin.GET("/user-sessions", aUser.GetOnlineUsers)
	admin.POST("/user-sessions/:id/force-offline", withFormID(aUser.ForceOfflineUser))
	admin.POST("/user-sessions/batch-force-offline", aUser.BatchForceOfflineUser)
}

func registerV2AdminContentRoutes(admin *route.RouterGroup) {
	aNews := adminnews.NewAdminNewsHandler()
	aEnroll := adminenroll.NewAdminEnrollHandler()
	aEvent := adminevent.NewAdminEventHandler()

	admin.GET("/news", aNews.GetAdminNewsList)
	admin.POST("/news", aNews.InsertNews)
	admin.GET("/news/:id", withQueryID(aNews.GetNewsDetail))
	admin.PUT("/news/:id", withFormID(aNews.EditNews))
	admin.DELETE("/news/:id", withFormID(aNews.DelNews))
	admin.DELETE("/news", aNews.DelNewsList)
	admin.PATCH("/news/:id/status", withFormID(aNews.StatusNews))
	admin.PATCH("/news/:id/recommendation", withFormID(aNews.VouchNews))
	admin.PATCH("/news/:id/sort", withFormID(aNews.SortNews))
	admin.PATCH("/news/:id/forms", withFormID(aNews.UpdateNewsForms))
	admin.PATCH("/news/:id/picture", withFormID(aNews.UpdateNewsPic))
	admin.PATCH("/news/:id/content", withFormID(aNews.UpdateNewsContent))

	admin.GET("/enrollments", aEnroll.GetAdminEnrollList)
	admin.POST("/enrollments", aEnroll.InsertEnroll)
	admin.GET("/enrollments/:id", withQueryID(aEnroll.GetEnrollDetail))
	admin.PUT("/enrollments/:id", withFormID(aEnroll.EditEnroll))
	admin.DELETE("/enrollments/:id", withFormID(aEnroll.DelEnroll))
	admin.DELETE("/enrollments", aEnroll.DelEnrolls)
	admin.PATCH("/enrollments/:id/status", withFormID(aEnroll.StatusEnroll))
	admin.PATCH("/enrollments/:id/sort", withFormID(aEnroll.SortEnroll))
	admin.PATCH("/enrollments/:id/recommendation", withFormID(aEnroll.VouchEnroll))
	admin.PATCH("/enrollments/:id/forms", withFormID(aEnroll.UpdateEnrollForms))
	admin.POST("/enrollments/:id/clear", withFormID(aEnroll.ClearEnrollAll))
	admin.GET("/enrollments/:id/joins", withQueryParam("enrollId", "id", aEnroll.GetEnrollJoinList))
	admin.DELETE("/enrollments/:id/joins/:joinId", withFormParams(map[string]string{"enrollJoinId": "joinId"}, aEnroll.DelEnrollJoin))
	admin.DELETE("/enrollments/:id/joins", aEnroll.DelEnrollJoins)
	admin.GET("/enrollments/:id/users", withQueryParam("enrollId", "id", aEnroll.GetEnrollUserList))
	admin.GET("/enrollments/:id/stats", withQueryParam("enrollId", "id", aEnroll.GetEnrollStats))
	admin.DELETE("/enrollments/:id/users/:userId", withFormParams(map[string]string{"enrollId": "id", "userId": "userId"}, aEnroll.RemoveEnrollUser))
	admin.DELETE("/enrollments/:id/users", withFormParam("enrollId", "id", aEnroll.RemoveEnrollUsers))
	admin.PUT("/enrollments/:id/users/:userId/forms", withFormParams(map[string]string{"enrollId": "id", "userId": "userId"}, aEnroll.EditEnrollUserForms))
	admin.GET("/enrollments/:id/export", withQueryParam("enrollId", "id", aEnroll.EnrollJoinDataGet))
	admin.POST("/enrollments/:id/export", withQueryParam("enrollId", "id", aEnroll.EnrollJoinDataExport))
	admin.DELETE("/enrollments/:id/export", withFormParam("enrollId", "id", aEnroll.EnrollJoinDataDel))

	admin.GET("/events", aEvent.GetAdminEventList)
	admin.POST("/events", aEvent.InsertEvent)
	admin.GET("/events/:id", withQueryID(aEvent.GetAdminEventDetail))
	admin.PUT("/events/:id", withFormID(aEvent.EditEvent))
	admin.DELETE("/events/:id", withFormID(aEvent.DelEvent))
	admin.DELETE("/events", aEvent.DelEvents)
	admin.PATCH("/events/:id/status", withFormID(aEvent.StatusEvent))
	admin.PATCH("/events/:id/recommendation", withFormID(aEvent.VouchEvent))
	admin.PATCH("/events/:id/top", withFormID(aEvent.TopEvent))
	admin.GET("/events/:id/participants", withQueryParam("eventId", "id", aEvent.GetEventParticipantList))
	admin.PUT("/events/:id/participants/:participantId", withFormParam("id", "participantId", aEvent.EditEventParticipant))
	admin.DELETE("/events/:id/participants/:participantId", withFormParam("id", "participantId", aEvent.DelEventParticipant))
	admin.DELETE("/events/:id/participants", aEvent.DelEventParticipants)
	admin.GET("/events/:id/dynamics", withQueryParam("eventId", "id", aEvent.GetEventDynamics))
	admin.POST("/events/:id/dynamics", withFormParam("eventId", "id", aEvent.PostEventDynamic))
	admin.PUT("/events/:id/dynamics/:dynamicId", withFormParam("id", "dynamicId", aEvent.EditEventDynamic))
	admin.DELETE("/events/:id/dynamics/:dynamicId", withFormParam("id", "dynamicId", aEvent.DelEventDynamic))
	admin.DELETE("/events/:id/dynamics", aEvent.DelEventDynamics)
	admin.GET("/events/:id/scores", withQueryParam("eventId", "id", aEvent.GetEventScores))
	admin.POST("/events/:id/scores", withFormParam("eventId", "id", aEvent.EditEventScore))
	admin.PUT("/events/:id/scores/:scoreId", withFormParam("id", "scoreId", aEvent.EditEventScore))
	admin.GET("/event-dept-users", aEvent.GetDeptUsers)
}

func registerV2AdminSystemRoutes(admin *route.RouterGroup) {
	aDict := admindict.NewAdminDictHandler()
	aDept := admindept.NewAdminDeptHandler()
	aRole := adminrole.NewAdminRoleHandler()
	aMenu := adminmenu.NewAdminMenuHandler()
	aPermission := adminpermission.NewAdminPermissionHandler()

	admin.GET("/dict/types", aDict.GetDictTypes)
	admin.GET("/dict/items", aDict.GetDictByType)
	admin.POST("/dict/items", aDict.AddDictItem)
	admin.PUT("/dict/items/:id", withFormID(aDict.EditDictItem))
	admin.DELETE("/dict/items/:id", withFormID(aDict.DelDictItem))
	admin.DELETE("/dict/types/:typeCode/items", withFormParam("typeCode", "typeCode", aDict.DelDictByType))
	admin.PATCH("/dict/types/:typeCode", withFormParam("oldTypeCode", "typeCode", aDict.EditDictTypeName))

	admin.GET("/departments/tree", aDept.GetDeptTree)
	admin.POST("/departments", aDept.AddDept)
	admin.PUT("/departments/:id", withFormID(aDept.EditDept))
	admin.DELETE("/departments/:id", withFormID(aDept.DelDept))

	admin.GET("/roles/application-permissions", aRole.GetApplicationPermissionTree)
	admin.GET("/roles", aRole.GetRoleList)
	admin.POST("/roles", aRole.AddRole)
	admin.PUT("/roles/:id", withFormID(aRole.EditRole))
	admin.DELETE("/roles/:id", withFormID(aRole.DelRole))
	admin.DELETE("/roles", aRole.DelRoles)

	admin.GET("/me/menus", aMenu.GetAdminMenus)
	admin.GET("/me/perms", aMenu.GetAdminPerms)

	admin.GET("/permissions/tree", aPermission.GetPermissionTree)
	admin.GET("/permissions", aPermission.GetPermissionList)
	admin.POST("/permissions", aPermission.AddPermission)
	admin.PUT("/permissions/:key", aPermission.EditPermission)
	admin.DELETE("/permissions/:key", withFormParam("key", "key", aPermission.DelPermission))
}

func registerV2AdminSurveyRoutes(admin *route.RouterGroup) {
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
	admin.GET("/surveys/:id", withQueryID(aSurvey.Detail))
	admin.PUT("/surveys/:id", withBodyOrFormID(aSurvey.Edit))
	admin.DELETE("/surveys/:id", withFormID(aSurvey.Del))
	admin.PATCH("/surveys/:id/status", withFormID(aSurvey.Status))
	admin.POST("/surveys/:id/copy", withFormID(aSurvey.Copy))
	admin.GET("/surveys/:id/statistics", withQueryParam("surveyId", "id", aSurvey.Statistic))
	admin.GET("/surveys/:id/responses", withQueryParam("surveyId", "id", aSurvey.ResponseList))
	admin.GET("/surveys/:id/responses/export", withQueryParam("surveyId", "id", aSurvey.ResponseExport))
	admin.GET("/survey-responses/:id", withQueryID(aSurvey.ResponseDetail))
	admin.DELETE("/survey-responses/:id", withFormID(aSurvey.ResponseDel))
	admin.DELETE("/survey-responses", aSurvey.ResponseBatchDel)
	admin.GET("/surveys/:id/channels", withQueryParam("surveyId", "id", aSurvey.ChannelList))
	admin.POST("/surveys/:id/channels", withBodyOrFormParam("surveyId", "id", aSurvey.ChannelInsert))
	admin.DELETE("/survey-channels/:id", withFormID(aSurvey.ChannelDel))
	admin.POST("/survey-resources", aSurvey.ResourceUpload)
	admin.GET("/surveys/:id/resources", withQueryParam("surveyId", "id", aSurvey.ResourceList))
	admin.DELETE("/survey-resources/:id", withFormID(aSurvey.ResourceDelete))
	admin.GET("/survey-question-bank", aSurvey.QuestionBankList)
	admin.POST("/survey-question-bank", aSurvey.QuestionBankInsert)
	admin.PUT("/survey-question-bank/:id", withBodyOrFormID(aSurvey.QuestionBankEdit))
	admin.DELETE("/survey-question-bank/:id", withFormID(aSurvey.QuestionBankDel))
	admin.GET("/survey-question-bank/categories", aSurvey.QuestionBankCategories)
	admin.GET("/survey-notifications", aSurvey.NotifyList)
	admin.PATCH("/survey-notifications/:id/read", withFormID(aSurvey.NotifyRead))
	admin.GET("/survey-notifications/unread-count", aSurvey.NotifyUnreadCount)
	admin.GET("/survey-template-presets", aSurvey.TemplatePresetsGet)
	admin.PUT("/survey-template-presets", aSurvey.TemplatePresetsSave)
}

func registerV2AdminExamRoutes(admin *route.RouterGroup) {
	aExam := adminexam.NewAdminExamHandler()

	admin.GET("/exams", aExam.List)
	admin.POST("/exams", aExam.Save)
	admin.GET("/exams/:id", withQueryID(aExam.Detail))
	admin.PUT("/exams/:id", withBodyOrFormID(aExam.Save))
	admin.DELETE("/exams/:id", withFormID(aExam.Delete))
	admin.PATCH("/exams/:id/status", withFormID(aExam.Status))
	admin.GET("/exams/:id/records", withQueryParam("examId", "id", aExam.RecordList))
	admin.GET("/exams/:id/records/:recordId", withQueryParam("id", "recordId", aExam.RecordDetail))
	admin.DELETE("/exams/:id/records/:recordId", withFormParam("id", "recordId", aExam.RecordDel))
	admin.DELETE("/exams/:id/records", aExam.RecordBatchDel)
	admin.GET("/exams/:id/statistics", withQueryParam("examId", "id", aExam.Statistics))
	admin.POST("/exam-resources", aExam.ResourceUpload)
	admin.GET("/exams/:id/resources", withQueryParam("examId", "id", aExam.ResourceList))
	admin.DELETE("/exam-resources/:id", withFormID(aExam.ResourceDelete))
	admin.GET("/exam-question-bank", aExam.QuestionBankList)
	admin.POST("/exam-question-bank", aExam.QuestionBankInsert)
	admin.PUT("/exam-question-bank/:id", withBodyOrFormID(aExam.QuestionBankEdit))
	admin.DELETE("/exam-question-bank/:id", withFormID(aExam.QuestionBankDel))
	admin.GET("/exam-question-bank/categories", aExam.QuestionBankCategories)
}

func withQueryID(next app.HandlerFunc) app.HandlerFunc {
	return withQueryParam("id", "id", next)
}

func withFormID(next app.HandlerFunc) app.HandlerFunc {
	return withFormParam("id", "id", next)
}

func withQueryParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if value := c.Param(routeParam); value != "" {
			c.QueryArgs().Set(name, value)
		}
		next(ctx, c)
	}
}

func withFormParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if value := c.Param(routeParam); value != "" {
			c.PostArgs().Set(name, value)
		}
		next(ctx, c)
	}
}

func withFormParams(params map[string]string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		for name, routeParam := range params {
			if value := c.Param(routeParam); value != "" {
				c.PostArgs().Set(name, value)
			}
		}
		next(ctx, c)
	}
}

func withBodyOrFormID(next app.HandlerFunc) app.HandlerFunc {
	return withBodyOrFormParam("id", "id", next)
}

func withBodyOrFormParam(name, routeParam string, next app.HandlerFunc) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		value := c.Param(routeParam)
		if value != "" {
			c.QueryArgs().Set(name, value)
			c.PostArgs().Set(name, value)
			injectJSONParam(c, name, value)
		}
		next(ctx, c)
	}
}

func injectJSONParam(c *app.RequestContext, name, value string) {
	contentType := strings.ToLower(string(c.Request.Header.ContentType()))
	if !strings.Contains(contentType, "json") {
		return
	}
	raw, err := c.Body()
	if err != nil || len(strings.TrimSpace(string(raw))) == 0 {
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}
	if n, err := strconv.ParseUint(value, 10, 64); err == nil {
		payload[name] = n
	} else {
		payload[name] = value
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	c.Request.SetBody(body)
}
