package review

import (
	"encoding/json"
	"strings"

	"wecheckin/backend/internal/model"
	templatesvc "wecheckin/backend/internal/service/dingtalkh5/performance/template"
	usersvc "wecheckin/backend/internal/service/dingtalkh5/performance/user"
)

const (
	ReviewStatusDraft           = "draft"
	ReviewStatusManagerReview   = "manager_review"
	ReviewStatusHRBPReview      = "hrbp_review"
	ReviewStatusEmployeeConfirm = "employee_confirm"
	ReviewStatusHRFinal         = "hr_final"
	ReviewStatusCompleted       = "completed"

	TemplateKeyDefault = "default"
)

type UserDTO = usersvc.UserDTO

type Objective struct {
	ID         string      `json:"id"`
	Target     string      `json:"target"`
	Weight     float64     `json:"weight"`
	Completion interface{} `json:"completion,omitempty"`
	Result     string      `json:"result,omitempty"`
}

type NextObjective = templatesvc.NextObjective

type ValueScore struct {
	ID         string        `json:"id"`
	Name       string        `json:"name,omitempty"`
	Definition string        `json:"definition,omitempty"`
	Rubric     []ValueRubric `json:"rubric,omitempty"`
	Self       interface{}   `json:"self"`
	Manager    interface{}   `json:"manager"`
	HRBP       interface{}   `json:"hrbp"`
	HR         interface{}   `json:"hr"`
}

type HistoryDTO struct {
	At     int64  `json:"at"`
	By     string `json:"by"`
	Action string `json:"action"`
}

type ReviewDTO struct {
	ID                      string          `json:"id"`
	DBID                    uint            `json:"dbId"`
	EmployeeID              string          `json:"employeeId"`
	EmployeeName            string          `json:"employeeName"`
	ManagerID               string          `json:"managerId"`
	ManagerName             string          `json:"managerName"`
	HRBPID                  string          `json:"hrbpId"`
	HRBPName                string          `json:"hrbpName"`
	HRBPReviewerID          string          `json:"hrbpReviewerId"`
	HRBPReviewerName        string          `json:"hrbpReviewerName"`
	Department              string          `json:"department"`
	DepartmentLevel1        string          `json:"departmentLevel1"`
	DepartmentLevel2        string          `json:"departmentLevel2"`
	DepartmentLevel3        string          `json:"departmentLevel3"`
	Period                  string          `json:"period"`
	NextPeriod              string          `json:"nextPeriod"`
	Status                  string          `json:"status"`
	ObjectiveSourceReviewID string          `json:"objectiveSourceReviewId"`
	ObjectiveSourcePeriod   string          `json:"objectiveSourcePeriod"`
	Objectives              []Objective     `json:"objectives"`
	NextObjectives          []NextObjective `json:"nextObjectives"`
	Values                  []ValueScore    `json:"values"`
	SelfSummary             string          `json:"selfSummary"`
	ManagerComment          string          `json:"managerComment"`
	ManagerGrade            string          `json:"managerGrade"`
	HRBPComment             string          `json:"hrbpComment"`
	HRBPGrade               string          `json:"hrbpGrade"`
	EmployeeConfirmResult   string          `json:"employeeConfirmResult"`
	EmployeeConfirmComment  string          `json:"employeeConfirmComment"`
	EmployeeConfirmedAt     int64           `json:"employeeConfirmedAt"`
	FinalGrade              string          `json:"finalGrade"`
	FinalNote               string          `json:"finalNote"`
	LatestAction            string          `json:"latestAction,omitempty"`
	History                 []HistoryDTO    `json:"history"`
}

type ReviewListResponse struct {
	List     []ReviewDTO `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type CreateReviewFailure struct {
	EmployeeID string `json:"employeeId"`
	Message    string `json:"message"`
}

type CreateReviewBatchResponse struct {
	List   []ReviewDTO           `json:"list"`
	Total  int                   `json:"total"`
	Failed []CreateReviewFailure `json:"failed,omitempty"`
}

type DingTalkH5Review struct {
	EmployeeID string
	ManagerID  string
	HRBPID     string
	Status     string
}

type GradeLevel = templatesvc.GradeLevel
type ValueRubric = templatesvc.ValueRubric
type ValueTemplate = templatesvc.ValueTemplate
type TemplateDTO = templatesvc.TemplateDTO

type ReviewPayload struct {
	EmployeeID             string          `json:"employeeId"`
	EmployeeIDs            []string        `json:"employeeIds"`
	Period                 string          `json:"period"`
	NextPeriod             string          `json:"nextPeriod"`
	Objectives             []Objective     `json:"objectives"`
	NextObjectives         []NextObjective `json:"nextObjectives"`
	Values                 []ValueScore    `json:"values"`
	SelfSummary            string          `json:"selfSummary"`
	ManagerComment         string          `json:"managerComment"`
	ManagerGrade           string          `json:"managerGrade"`
	HRBPComment            string          `json:"hrbpComment"`
	HRBPGrade              string          `json:"hrbpGrade"`
	EmployeeConfirmComment string          `json:"employeeConfirmComment"`
	FinalGrade             string          `json:"finalGrade"`
	FinalNote              string          `json:"finalNote"`
	ReturnReason           string          `json:"returnReason"`
}

type UserPayload = usersvc.UserPayload

type ReviewFilters struct {
	Keyword         string
	Scope           string
	EmployeeName    string
	Department      string
	DepartmentName  string
	DepartmentNames []string
	Period          string
	Periods         []string
	Year            string
	Month           string
	NotPeriod       string
	NextPeriod      string
	Status          string
	Statuses        []string
	ManagerID       string
	HRBPID          string
	Grade           string
	ObjectiveScore  string
	Page            int
	PageSize        int
	SkipHistory     bool
	Detail          bool
}

func userDTO(user model.DingTalkH5PerfUser) UserDTO {
	return usersvc.UserDTOFromModel(user)
}

func reviewDTO(review model.DingTalkH5PerfReview, histories []model.DingTalkH5PerfHistory) ReviewDTO {
	result := ReviewDTO{
		ID:                      review.ReviewNo,
		DBID:                    review.ID,
		EmployeeID:              review.EmployeeAccount,
		ManagerID:               review.ManagerAccount,
		HRBPID:                  review.HRBPAccount,
		HRBPReviewerID:          review.HRBPReviewerAccount,
		Department:              review.Department,
		DepartmentLevel1:        review.DepartmentLevel1,
		DepartmentLevel2:        review.DepartmentLevel2,
		DepartmentLevel3:        review.DepartmentLevel3,
		Period:                  review.Period,
		NextPeriod:              review.NextPeriod,
		Status:                  review.Status,
		ObjectiveSourceReviewID: review.ObjectiveSourceReviewNo,
		ObjectiveSourcePeriod:   review.ObjectiveSourcePeriod,
		Objectives:              decodeObjectives(review.ObjectivesJSON),
		NextObjectives:          decodeNextObjectives(review.NextObjectivesJSON),
		Values:                  decodeValues(review.ValuesJSON),
		SelfSummary:             review.SelfSummary,
		ManagerComment:          review.ManagerComment,
		ManagerGrade:            review.ManagerGrade,
		HRBPComment:             review.HRBPComment,
		HRBPGrade:               review.HRBPGrade,
		EmployeeConfirmResult:   review.EmployeeConfirmResult,
		EmployeeConfirmComment:  review.EmployeeConfirmComment,
		EmployeeConfirmedAt:     review.EmployeeConfirmedAt,
		FinalGrade:              review.FinalGrade,
		FinalNote:               review.FinalNote,
	}
	for _, item := range histories {
		result.History = append(result.History, HistoryDTO{At: item.AddTime, By: item.ByName, Action: item.Action})
	}
	if len(histories) > 0 {
		result.LatestAction = histories[len(histories)-1].Action
	}
	return result
}

func reviewDTOWithUsers(review model.DingTalkH5PerfReview, histories []model.DingTalkH5PerfHistory, users map[string]*model.DingTalkH5PerfUser) ReviewDTO {
	result := reviewDTO(review, histories)
	result.EmployeeName = reviewUserName(users, review.EmployeeAccount)
	result.ManagerName = reviewUserName(users, review.ManagerAccount)
	result.HRBPName = reviewUserName(users, review.HRBPAccount)
	result.HRBPReviewerName = reviewUserName(users, review.HRBPReviewerAccount)
	return result
}

func reviewUserName(users map[string]*model.DingTalkH5PerfUser, account string) string {
	if users == nil || account == "" {
		return ""
	}
	user := users[account]
	if user == nil {
		return ""
	}
	return strings.TrimSpace(user.Name)
}

func encodeJSON(value interface{}) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func decodeObjectives(raw string) []Objective {
	var items []Objective
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func decodeNextObjectives(raw string) []NextObjective {
	var items []NextObjective
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func decodeValues(raw string) []ValueScore {
	var items []ValueScore
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func decodeStringList(raw string) []string {
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	return items
}

func toString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		data, _ := json.Marshal(typed)
		return strings.Trim(string(data), `"`)
	}
}
