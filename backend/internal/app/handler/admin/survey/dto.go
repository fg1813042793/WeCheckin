package survey

import (
	"wecheckin/backend/internal/app/formkit/report"
	"wecheckin/backend/internal/model"
)

type surveyListItem struct {
	ID            uint   `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	Tags          string `json:"tags"`
	Cover         string `json:"cover"`
	Visibility    int    `json:"visibility"`
	AllowMulti    int    `json:"allowMulti"`
	StartTime     int64  `json:"startTime"`
	EndTime       int64  `json:"endTime"`
	MaxResponse   int    `json:"maxResponse"`
	ShowResult    int    `json:"showResult"`
	Anonymous     int    `json:"anonymous"`
	DeptIDs       string `json:"deptIds"`
	QR            string `json:"qr"`
	Status        int    `json:"status"`
	Mode          string `json:"mode"`
	Order         int    `json:"order"`
	DeptID        uint   `json:"deptId"`
	CreateBy      uint   `json:"createBy"`
	AddTime       int64  `json:"addTime"`
	EditTime      int64  `json:"editTime"`
	ResponseCount int    `json:"responseCount"`
}

type surveyListResponse struct {
	List  []surveyListItem `json:"list"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
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

func newSurveyListItem(sv model.Survey, responseCount int) surveyListItem {
	return surveyListItem{
		ID:            sv.ID,
		Title:         sv.Title,
		Description:   sv.Desc,
		Category:      sv.Category,
		Tags:          sv.Tags,
		Cover:         sv.Cover,
		Visibility:    sv.Visibility,
		AllowMulti:    sv.AllowMulti,
		StartTime:     sv.StartTime,
		EndTime:       sv.EndTime,
		MaxResponse:   sv.MaxResponse,
		ShowResult:    sv.ShowResult,
		Anonymous:     sv.Anonymous,
		DeptIDs:       sv.DeptIDs,
		QR:            sv.QR,
		Status:        sv.Status,
		Mode:          sv.Mode,
		Order:         sv.Order,
		DeptID:        sv.DeptID,
		CreateBy:      sv.CreateBy,
		AddTime:       sv.AddTime,
		EditTime:      sv.EditTime,
		ResponseCount: responseCount,
	}
}
