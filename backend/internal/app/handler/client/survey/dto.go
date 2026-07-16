package survey

import "wecheckin-backend/backend/internal/model"

type limitInfo struct {
	DeviceFull bool `json:"deviceFull"`
	IPFull     bool `json:"ipFull"`
}

type listResponse struct {
	List   []model.Survey     `json:"list"`
	Total  int64              `json:"total"`
	Page   int                `json:"page"`
	Size   int                `json:"size"`
	Limits map[uint]limitInfo `json:"limits"`
}

type detailResponse struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Cover       string                 `json:"cover"`
	Visibility  int                    `json:"visibility"`
	Anonymous   int                    `json:"anonymous"`
	AllowMulti  int                    `json:"allowMulti"`
	StartTime   int64                  `json:"startTime"`
	EndTime     int64                  `json:"endTime"`
	MaxResponse int                    `json:"maxResponse"`
	ShowResult  int                    `json:"showResult"`
	Schema      map[string]interface{} `json:"schema"`
	Settings    map[string]interface{} `json:"settings"`
	Session     string                 `json:"session"`
	StartAt     int64                  `json:"startAt"`
	DeptIDs     string                 `json:"deptIds"`
}

type answersResponse struct {
	Answers map[string]interface{} `json:"answers"`
}

type validationResponse struct {
	Errors interface{} `json:"errors"`
	Valid  bool        `json:"valid"`
}

type submitResponse struct {
	ID         uint  `json:"id"`
	SubmitTime int64 `json:"submitTime"`
}

type myResponsesResponse struct {
	List []model.SurveyResponse `json:"list"`
}

type myResponseDetailResponse struct {
	Response *model.SurveyResponse  `json:"response"`
	Answers  map[string]interface{} `json:"answers"`
	Survey   *model.Survey          `json:"survey,omitempty"`
}

type publicValidationResponse struct {
	Valid  bool        `json:"valid"`
	Errors interface{} `json:"errors"`
}

type publicApplyResponse struct {
	Answers map[string]interface{} `json:"answers"`
	States  interface{}            `json:"states"`
}
