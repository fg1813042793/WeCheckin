package exam

import (
	"errors"
	"wecheckin/backend/internal/model"
)

var (
	ErrRecordNotGradable        = errors.New("答卷当前状态不允许批改")
	ErrRecordGradingChanged     = errors.New("答卷或评分资料已更新，请重新打开批改页面")
	ErrRecordGradingInvalid     = errors.New("请为全部待批改题目填写有效整数分数，且不得超过该题满分；评语最多1000字")
	ErrRecordGradingUnavailable = errors.New("答卷缺少有效的逐题评分资料，无法人工批改")
)

type ManualQuestionScore struct {
	QuestionID uint `json:"questionId"`
	Score      *int `json:"score"`
}

type GradeRecordRequest struct {
	ExpectedVersion string                `json:"expectedVersion"`
	Scores          []ManualQuestionScore `json:"scores"`
	Comment         string                `json:"comment"`
}

type GradingQuestion struct {
	QuestionID      uint              `json:"questionId"`
	Title           string            `json:"title"`
	Type            string            `json:"type"`
	Answer          string            `json:"answer"`
	ReferenceAnswer string            `json:"referenceAnswer"`
	FullScore       int               `json:"fullScore"`
	GotScore        int               `json:"gotScore"`
	NeedManual      bool              `json:"needManual"`
	Unscored        bool              `json:"unscored"` // 仅展示题目与答案，不参与评分。
	Materials       []GradingMaterial `json:"materials,omitempty"`
	Images          []GradingImage    `json:"images,omitempty"`
}

type GradingMaterial struct {
	Label   string `json:"label"`
	Content string `json:"content"`
}
type GradingImage struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

type RecordGradingDetail struct {
	Record    model.ExamRecord  `json:"record"`
	ExamTitle string            `json:"examTitle"`
	PassScore int               `json:"passScore"`
	Version   string            `json:"version"`
	CanGrade  bool              `json:"canGrade"`
	Notice    string            `json:"notice"`
	Questions []GradingQuestion `json:"questions"`
}
