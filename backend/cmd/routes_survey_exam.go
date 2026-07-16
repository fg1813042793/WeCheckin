package main

import (
	"github.com/cloudwego/hertz/pkg/route"
	adminexam "wecheckin-backend/backend/internal/app/handler/admin/exam"
	adminsurvey "wecheckin-backend/backend/internal/app/handler/admin/survey"
)

func registerAdminSurveyRoutes(adminGroup *route.RouterGroup) {
	aSurvey := adminsurvey.NewAdminSurveyHandler()

	adminGroup.GET("/survey/types", aSurvey.ListTypes)
	adminGroup.POST("/survey/schema/parse", aSurvey.ParseSchema)
	adminGroup.POST("/survey/eval", aSurvey.EvalExpr)

	adminGroup.GET("/survey/report/enroll", aSurvey.ReportEnrollSchema)
	adminGroup.GET("/survey/export/enroll", aSurvey.ExportEnrollSchemaCSV)
	adminGroup.GET("/survey/report/event", aSurvey.ReportEventSchema)
	adminGroup.GET("/survey/export/event", aSurvey.ExportEventSchemaCSV)
	adminGroup.GET("/survey/report/survey", aSurvey.ReportSurveySchema)
	adminGroup.GET("/survey/export/survey", aSurvey.ExportSurveySchemaCSV)

	adminGroup.GET("/survey/survey_list", aSurvey.List)
	adminGroup.GET("/survey/survey_detail", aSurvey.Detail)
	adminGroup.POST("/survey/survey_insert", aSurvey.Insert)
	adminGroup.POST("/survey/survey_edit", aSurvey.Edit)
	adminGroup.POST("/survey/survey_del", aSurvey.Del)
	adminGroup.POST("/survey/survey_status", aSurvey.Status)
	adminGroup.POST("/survey/survey_copy", aSurvey.Copy)
	adminGroup.GET("/survey/response_list", aSurvey.ResponseList)
	adminGroup.GET("/survey/response_detail", aSurvey.ResponseDetail)
	adminGroup.POST("/survey/response_del", aSurvey.ResponseDel)
	adminGroup.POST("/survey/response_batch_del", aSurvey.ResponseBatchDel)
	adminGroup.GET("/survey/response_export", aSurvey.ResponseExport)
	adminGroup.GET("/survey/statistic", aSurvey.Statistic)
	adminGroup.GET("/survey/channel_list", aSurvey.ChannelList)
	adminGroup.POST("/survey/channel_insert", aSurvey.ChannelInsert)
	adminGroup.POST("/survey/channel_del", aSurvey.ChannelDel)
	adminGroup.POST("/survey/resource_upload", aSurvey.ResourceUpload)
	adminGroup.GET("/survey/resource_list", aSurvey.ResourceList)
	adminGroup.POST("/survey/resource_delete", aSurvey.ResourceDelete)
	adminGroup.GET("/survey/question_bank_list", aSurvey.QuestionBankList)
	adminGroup.POST("/survey/question_bank_insert", aSurvey.QuestionBankInsert)
	adminGroup.POST("/survey/question_bank_edit", aSurvey.QuestionBankEdit)
	adminGroup.POST("/survey/question_bank_del", aSurvey.QuestionBankDel)
	adminGroup.GET("/survey/question_bank_categories", aSurvey.QuestionBankCategories)
	adminGroup.GET("/survey/notify_list", aSurvey.NotifyList)
	adminGroup.POST("/survey/notify_read", aSurvey.NotifyRead)
	adminGroup.GET("/survey/notify_unread_count", aSurvey.NotifyUnreadCount)
	adminGroup.GET("/survey/template_presets", aSurvey.TemplatePresetsGet)
	adminGroup.POST("/survey/template_presets", aSurvey.TemplatePresetsSave)
}

func registerAdminExamRoutes(adminGroup *route.RouterGroup) {
	aExam := adminexam.NewAdminExamHandler()

	adminGroup.GET("/exam/list", aExam.List)
	adminGroup.GET("/exam/detail", aExam.Detail)
	adminGroup.POST("/exam/save", aExam.Save)
	adminGroup.POST("/exam/status", aExam.Status)
	adminGroup.POST("/exam/delete", aExam.Delete)
	adminGroup.GET("/exam/record/list", aExam.RecordList)
	adminGroup.GET("/exam/record/detail", aExam.RecordDetail)
	adminGroup.POST("/exam/record/del", aExam.RecordDel)
	adminGroup.POST("/exam/record/batch_del", aExam.RecordBatchDel)
	adminGroup.GET("/exam/statistics", aExam.Statistics)
	adminGroup.POST("/exam/resource_upload", aExam.ResourceUpload)
	adminGroup.GET("/exam/resource_list", aExam.ResourceList)
	adminGroup.POST("/exam/resource_delete", aExam.ResourceDelete)
	adminGroup.GET("/exam/question_bank_list", aExam.QuestionBankList)
	adminGroup.POST("/exam/question_bank_insert", aExam.QuestionBankInsert)
	adminGroup.POST("/exam/question_bank_edit", aExam.QuestionBankEdit)
	adminGroup.POST("/exam/question_bank_del", aExam.QuestionBankDel)
	adminGroup.GET("/exam/question_bank_categories", aExam.QuestionBankCategories)
}
