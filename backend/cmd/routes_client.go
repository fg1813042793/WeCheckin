package main

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	clientenroll "wecheckin/backend/internal/handler/client/enroll"
	clientevent "wecheckin/backend/internal/handler/client/event"
	clientexam "wecheckin/backend/internal/handler/client/exam"
	clientfavorite "wecheckin/backend/internal/handler/client/favorite"
	clientnews "wecheckin/backend/internal/handler/client/news"
	clientpassport "wecheckin/backend/internal/handler/client/passport"
	clientsurvey "wecheckin/backend/internal/handler/client/survey"
	clientmw "wecheckin/backend/internal/middleware/client"
)

func registerClientRoutes(h *server.Hertz) {
	pp := clientpassport.NewPassportHandler()
	ns := clientnews.NewNewsHandler()
	el := clientenroll.NewEnrollHandler()
	fa := clientfavorite.NewFavHandler()
	ev := clientevent.NewEventHandler()
	cSurvey := clientsurvey.NewClientSurveyHandler()
	cExam := clientexam.NewClientExamHandler()

	clientGroup := h.Group("/passport", clientmw.ClientAuth(), clientmw.ClientPerm())
	clientGroup.POST("/phone", pp.GetPhone)
	clientGroup.GET("/my_detail", pp.GetMyDetail)
	clientGroup.POST("/edit_base", pp.EditBase)
	clientGroup.POST("/logout", pp.Logout)

	clientFav := h.Group("/fav", clientmw.ClientAuth(), clientmw.ClientPerm())
	clientFav.POST("/update", fa.UpdateFav)
	clientFav.POST("/del", fa.DelFav)
	clientFav.GET("/is_fav", fa.IsFav)
	clientFav.GET("/my_list", fa.GetMyFavList)

	clientNews := h.Group("/news", clientmw.ClientAuth(), clientmw.ClientPerm())
	clientNews.GET("/list", ns.GetNewsList)
	clientNews.GET("/view", ns.ViewNews)
	clientNews.GET("/cate_list", ns.GetNewsCateList)

	clientEnroll := h.Group("/enroll", clientmw.ClientAuth(), clientmw.ClientPerm())
	clientEnroll.GET("/list", el.GetEnrollList)
	clientEnroll.GET("/view", el.ViewEnroll)
	clientEnroll.GET("/join_day", el.GetEnrollJoinByDay)
	clientEnroll.POST("/join", el.EnrollJoin)
	clientEnroll.POST("/enroll_submit", el.EnrollUserSubmit)
	clientEnroll.GET("/my_join_list", el.GetMyEnrollJoinList)
	clientEnroll.GET("/my_user_list", el.GetMyEnrollUserList)
	clientEnroll.GET("/my_records", el.GetMyJoinRecords)
	clientEnroll.GET("/my_calendar", el.GetMyCalendar)
	clientEnroll.GET("/my_day_records", el.GetMyDayRecords)

	clientEvent := h.Group("/event")
	clientEvent.GET("/list", ev.GetEventList)
	clientEvent.GET("/view", ev.ViewEvent)

	clientEventAuth := h.Group("/event", clientmw.ClientAuth(), clientmw.ClientPerm())
	clientEventAuth.POST("/participate", ev.EventParticipate)
	clientEventAuth.GET("/my_list", ev.GetMyEventList)
	clientEventAuth.GET("/my_roles", ev.GetMyEventRoles)
	clientEventAuth.GET("/my_managed", ev.GetMyManagedList)
	clientEventAuth.GET("/dynamics", ev.GetEventDynamics)
	clientEventAuth.POST("/dynamic_post", ev.PostEventDynamic)
	clientEventAuth.GET("/participant_list", ev.GetEventParticipantList)
	clientEventAuth.GET("/scores", ev.GetEventScores)
	clientEventAuth.POST("/score_save", ev.SaveEventScore)

	surveyPub := h.Group("/survey")
	surveyPub.GET("/list", cSurvey.List)
	surveyPub.GET("/view", cSurvey.Detail)
	surveyPub.POST("/submit", cSurvey.Submit)
	surveyAuth := h.Group("/survey", clientmw.ClientAuth(), clientmw.ClientPerm())
	surveyAuth.GET("/my_responses", cSurvey.MyResponses)
	surveyAuth.GET("/my_response", cSurvey.MyResponseDetail)

	examPub := h.Group("/exam")
	examPub.GET("/list", cExam.List)
	examPub.GET("/view", cExam.View)
	examPub.POST("/submit", cExam.Submit)
	examPub.POST("/validate", cExam.Validate)
	examPub.GET("/result", cExam.ResultBySession)
	examAuth := h.Group("/exam", clientmw.ClientAuth(), clientmw.ClientPerm())
	examAuth.GET("/start", cExam.Start)
	examAuth.POST("/save_answer", cExam.SaveAnswer)
	examAuth.GET("/record", cExam.Record)
	examAuth.GET("/my_records", cExam.MyRecords)
}
