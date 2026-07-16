package survey

import (
	"wecheckin-backend/backend/internal/app/formkit/report"
	"wecheckin-backend/backend/internal/model"
)

type surveyWithCount struct {
	model.Survey
	ResponseCount int `json:"responseCount"`
}

type surveyListResponse struct {
	List  []surveyWithCount `json:"list"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Size  int               `json:"size"`
}

type surveyDetailResponse struct {
	Survey        *model.Survey `json:"survey"`
	ResponseCount int64         `json:"responseCount"`
	Schema        string        `json:"schema"`
}

type surveyResponseWithAnswers struct {
	model.SurveyResponse
	AnswersMap map[string]interface{} `json:"answers"`
}

type surveyResponseListResponse struct {
	List  []surveyResponseWithAnswers `json:"list"`
	Total int64                       `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

type surveyResponseDetailResponse struct {
	Response *model.SurveyResponse  `json:"response"`
	Survey   *model.Survey          `json:"survey"`
	Answers  map[string]interface{} `json:"answers"`
	Schema   map[string]interface{} `json:"schema"`
}

type surveyChannelListResponse struct {
	List []model.SurveyChannel `json:"list"`
}

type surveyNotificationListResponse struct {
	List  []model.Notify `json:"list"`
	Total int64          `json:"total"`
}

type surveyQuestionListResponse struct {
	List  []model.SurveyQuestion `json:"list"`
	Total int64                  `json:"total"`
}

type surveyDailyCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type surveyDeviceStat struct {
	Mobile int64 `json:"mobile"`
	PC     int64 `json:"pc"`
}

type surveyStatisticResponse struct {
	Survey     *model.Survey      `json:"survey"`
	Total      int64              `json:"total"`
	TodayCount int64              `json:"todayCount"`
	Daily      []surveyDailyCount `json:"daily"`
	DeviceStat surveyDeviceStat   `json:"deviceStat"`
	FieldStats []report.FieldStat `json:"fieldStats"`
	ViewURL    string             `json:"viewUrl"`
}
