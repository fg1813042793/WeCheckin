package client

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	admindict "wecheckin/backend/internal/handler/admin/dict"
	adminuser "wecheckin/backend/internal/handler/admin/user"
	clientenroll "wecheckin/backend/internal/handler/client/enroll"
	clientevent "wecheckin/backend/internal/handler/client/event"
	clientexam "wecheckin/backend/internal/handler/client/exam"
	clientfavorite "wecheckin/backend/internal/handler/client/favorite"
	clientnews "wecheckin/backend/internal/handler/client/news"
	clientpassport "wecheckin/backend/internal/handler/client/passport"
	clientsurvey "wecheckin/backend/internal/handler/client/survey"
	publicgeo "wecheckin/backend/internal/handler/public/geo"
	publichome "wecheckin/backend/internal/handler/public/home"
	clientmw "wecheckin/backend/internal/middleware/client"
	workflowapp "wecheckin/backend/internal/modules/workflow/application"
	workflowinfra "wecheckin/backend/internal/modules/workflow/infrastructure"
	workflowhttp "wecheckin/backend/internal/modules/workflow/transport/httpclient"
	"wecheckin/backend/internal/routes/v2/routeparam"
	"wecheckin/backend/pkg/database"
)

func Register(h *server.Hertz) {
	registerPublicRoutes(h)
	registerAuthenticatedRoutes(h)
}

func registerPublicRoutes(h *server.Hertz) {
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
	h.GET("/api/v2/dict/types", aDict.GetPublicDictTypes)
	h.GET("/api/v2/dict/items", aDict.GetPublicDictByType)
	h.GET("/api/v2/events", ev.GetEventList)
	h.GET("/api/v2/events/:id", routeparam.WithQueryID(ev.ViewEvent))
	h.GET("/api/v2/surveys", cSurvey.List)
	h.GET("/api/v2/surveys/:id", routeparam.WithQueryID(cSurvey.Detail))
	h.POST("/api/v2/surveys/:id/responses", routeparam.WithBodyOrFormParam("surveyId", "id", cSurvey.Submit))
	h.POST("/api/v2/survey/apply", cSurvey.PublicApply)
	h.POST("/api/v2/survey/validate", cSurvey.PublicValidate)
	h.GET("/api/v2/exams", cExam.List)
	h.GET("/api/v2/exams/:id", routeparam.WithQueryID(cExam.View))
	h.POST("/api/v2/exams/:id/submissions", routeparam.WithBodyOrFormParam("examId", "id", cExam.Submit))
	h.POST("/api/v2/exams/:id/validation", routeparam.WithBodyOrFormParam("examId", "id", cExam.Validate))
	h.GET("/api/v2/exam-results", cExam.ResultBySession)
}

func registerAuthenticatedRoutes(h *server.Hertz) {
	pp := clientpassport.NewPassportHandler()
	ns := clientnews.NewNewsHandler()
	el := clientenroll.NewEnrollHandler()
	fa := clientfavorite.NewFavHandler()
	ev := clientevent.NewEventHandler()
	cSurvey := clientsurvey.NewClientSurveyHandler()
	cExam := clientexam.NewClientExamHandler()
	db := database.GetDB()
	workflowStore := workflowinfra.NewGormStore(db)
	notificationRepository := workflowinfra.NewGormNotificationRepository(db)
	notificationDispatcher := workflowapp.NewNotificationDispatcher(
		notificationRepository,
		workflowinfra.NewDingTalkNotificationChannel(db, nil),
	)
	workflowService := workflowapp.NewServiceWithNotifications(
		workflowStore,
		workflowinfra.NewAssigneeResolver(db),
		workflowinfra.NewRandomIDGenerator(),
		workflowapp.DefaultLifecycleEventPublisher(),
		notificationDispatcher,
	)
	workflowHandler := workflowhttp.NewRuntimeHandler(workflowService)

	client := h.Group("/api/v2", clientmw.ClientAuth(), clientmw.ClientPerm())
	client.GET("/me/bootstrap", pp.Bootstrap)
	client.GET("/me", pp.GetMyDetail)
	client.PUT("/me", pp.EditBase)
	client.POST("/me/phone", pp.GetPhone)
	client.POST("/me/logout", pp.Logout)
	client.GET("/me/favorites", fa.GetMyFavList)
	client.POST("/me/favorites", fa.UpdateFav)
	client.DELETE("/me/favorites/:oid", routeparam.WithFormParam("oid", "oid", fa.DelFav))
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
	client.GET("/me/survey-responses/:id", routeparam.WithQueryID(cSurvey.MyResponseDetail))
	client.GET("/me/exam-records", cExam.MyRecords)

	client.GET("/news", ns.GetNewsList)
	client.GET("/news/categories", ns.GetNewsCateList)
	client.GET("/news/:id", routeparam.WithQueryID(ns.ViewNews))
	client.GET("/enrollments", el.GetEnrollList)
	client.GET("/enrollments/:id", routeparam.WithQueryID(el.ViewEnroll))
	client.GET("/enrollments/:id/join-days", routeparam.WithQueryID(el.GetEnrollJoinByDay))
	client.POST("/enrollments/:id/joins", routeparam.WithFormParam("enroll_id", "id", el.EnrollJoin))
	client.POST("/enrollments/:id/submissions", routeparam.WithFormParam("enroll_id", "id", el.EnrollUserSubmit))
	client.POST("/events/:id/participants", routeparam.WithFormParam("event_id", "id", ev.EventParticipate))
	client.GET("/events/:id/participants", routeparam.WithQueryParam("event_id", "id", ev.GetEventParticipantList))
	client.GET("/events/:id/dynamics", routeparam.WithQueryParam("event_id", "id", ev.GetEventDynamics))
	client.POST("/events/:id/dynamics", routeparam.WithFormParam("event_id", "id", ev.PostEventDynamic))
	client.GET("/events/:id/scores", routeparam.WithQueryParam("event_id", "id", ev.GetEventScores))
	client.POST("/events/:id/scores", routeparam.WithFormParam("event_id", "id", ev.SaveEventScore))
	client.POST("/exams/:id/start", routeparam.WithQueryParam("examId", "id", cExam.Start))
	client.GET("/exam-records/:id", routeparam.WithQueryID(cExam.Record))
	client.PUT("/exam-records/:id/answers", routeparam.WithFormParam("recordId", "id", cExam.SaveAnswer))
	client.GET("/workflows/definitions", workflowHandler.ListDefinitions)
	client.GET("/workflows/definitions/:id", workflowHandler.GetDefinition)
	client.GET("/workflows/drafts/:definitionId", workflowHandler.GetStartDraft)
	client.PUT("/workflows/drafts/:definitionId", workflowHandler.SaveStartDraft)
	client.DELETE("/workflows/drafts/:definitionId", workflowHandler.DeleteStartDraft)
	client.POST("/workflows/instances", workflowHandler.StartInstance)
	client.GET("/workflows/instances", workflowHandler.ListMyInstances)
	client.GET("/workflows/instances/:id", workflowHandler.GetMyInstance)
	client.POST("/workflows/instances/:id/withdraw", workflowHandler.WithdrawInstance)
	client.GET("/workflows/tasks", workflowHandler.ListMyTasks)
	client.POST("/workflows/tasks/:id/complete", workflowHandler.CompleteTask)
}
