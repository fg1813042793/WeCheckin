package exam

import (
	examPkg "wecheckin/backend/internal/app/formkit/exam"
	examservice "wecheckin/backend/internal/app/service/exam"
	"wecheckin/backend/internal/model"
)

type examLimitInfo struct {
	DeviceFull bool `json:"deviceFull,omitempty"`
	IPFull     bool `json:"ipFull,omitempty"`
}

type examListResponse struct {
	List   []model.Exam           `json:"list"`
	Total  int64                  `json:"total"`
	Page   int                    `json:"page"`
	Size   int                    `json:"size"`
	Limits map[uint]examLimitInfo `json:"limits"`
}

type examNotStartedData struct {
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Title     string `json:"title"`
}

type examViewPaperResponse struct {
	Exam      model.Exam                     `json:"exam"`
	Paper     model.ExamPaper                `json:"paper"`
	Questions []examservice.SafeExamQuestion `json:"questions"`
	StartAt   int64                          `json:"startAt"`
	Session   string                         `json:"session"`
}

type examViewSchemaResponse struct {
	ID          uint                   `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Visibility  int                    `json:"visibility"`
	Anonymous   int                    `json:"anonymous"`
	ShowResult  int                    `json:"showResult"`
	ShowScore   int                    `json:"showScore"`
	Duration    int                    `json:"duration"`
	MaxAttempts int                    `json:"maxAttempts"`
	StartTime   int64                  `json:"startTime"`
	EndTime     int64                  `json:"endTime"`
	Schema      map[string]interface{} `json:"schema"`
	Settings    map[string]interface{} `json:"settings"`
	StartAt     int64                  `json:"startAt"`
	Session     string                 `json:"session"`
	DeptIDs     string                 `json:"deptIds"`
	Mode        string                 `json:"mode"`
}

type examStartResponse struct {
	Record    model.ExamRecord               `json:"record"`
	Paper     model.ExamPaper                `json:"paper"`
	Exam      model.Exam                     `json:"exam"`
	Questions []examservice.SafeExamQuestion `json:"questions"`
	Answers   map[string]interface{}         `json:"answers"`
}

type examRecordResponse struct {
	Record    model.ExamRecord               `json:"record"`
	Exam      model.Exam                     `json:"exam"`
	Paper     model.ExamPaper                `json:"paper"`
	Questions []examservice.SafeExamQuestion `json:"questions"`
	Answers   map[string]interface{}         `json:"answers"`
	Results   []examPkg.Result               `json:"results"`
}

type examMyRecordsResponse struct {
	List []model.ExamRecord `json:"list"`
}

type examResultBySessionResponse struct {
	Exam      model.Exam               `json:"exam"`
	Record    model.ExamRecord         `json:"record"`
	Questions []map[string]interface{} `json:"questions"`
	Answers   map[string]interface{}   `json:"answers"`
	Results   []examPkg.Result         `json:"results"`
	Settings  map[string]interface{}   `json:"settings"`
}

type examSubmitResponse struct {
	Score      int              `json:"score"`
	FullScore  int              `json:"fullScore"`
	CorrectCnt int              `json:"correctCnt"`
	ManualCnt  int              `json:"manualCnt"`
	Results    []examPkg.Result `json:"results"`
}

type examValidationResponse struct {
	OK     bool        `json:"ok"`
	Errors interface{} `json:"errors"`
}
