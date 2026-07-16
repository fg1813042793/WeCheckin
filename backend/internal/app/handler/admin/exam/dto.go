package exam

import "wecheckin-backend/backend/internal/model"

type examDetailSurveyDTO struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Tags        string `json:"tags"`
	Visibility  int    `json:"visibility"`
	AllowMulti  int    `json:"allowMulti"`
	Anonymous   int    `json:"anonymous"`
	ShowResult  int    `json:"showResult"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	MaxResponse int    `json:"maxResponse"`
	Duration    int    `json:"duration"`
	MaxAttempts int    `json:"maxAttempts"`
	ShowScore   int    `json:"showScore"`
	Status      int    `json:"status"`
	DeptIds     string `json:"deptIds"`
	Mode        string `json:"mode"`
	CreateBy    uint   `json:"createBy"`
	Settings    string `json:"settings"`
}

type examDetailResponse struct {
	Survey        examDetailSurveyDTO `json:"survey"`
	ResponseCount int64               `json:"responseCount"`
	Schema        string              `json:"schema"`
}

type examSaveResponse struct {
	ID uint `json:"id"`
}

type examListResponse struct {
	List  []model.Exam `json:"list"`
	Total int64        `json:"total"`
	Page  int          `json:"page"`
	Size  int          `json:"size"`
}

type examRecordListResponse struct {
	List  []model.ExamRecord `json:"list"`
	Total int64              `json:"total"`
}

func newExamDetailSurveyDTO(exam *model.Exam) examDetailSurveyDTO {
	return examDetailSurveyDTO{
		ID:          exam.ID,
		Title:       exam.Title,
		Description: exam.Description,
		Category:    exam.Category,
		Tags:        exam.Tags,
		Visibility:  exam.Visibility,
		AllowMulti:  exam.AllowMulti,
		Anonymous:   exam.Anonymous,
		ShowResult:  exam.ShowResult,
		StartTime:   exam.StartTime,
		EndTime:     exam.EndTime,
		MaxResponse: exam.MaxResponse,
		Duration:    exam.Duration,
		MaxAttempts: exam.MaxAttempts,
		ShowScore:   exam.ShowScore,
		Status:      exam.Status,
		DeptIds:     exam.DeptIds,
		Mode:        exam.Mode,
		CreateBy:    exam.CreateBy,
		Settings:    exam.Settings,
	}
}
