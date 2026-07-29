package dingtalkh5

import (
	"encoding/json"
	"strings"

	"wecheckin-backend/backend/internal/model"
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

type UserDTO struct {
	ID                     string   `json:"id"`
	Account                string   `json:"account"`
	Name                   string   `json:"name"`
	Role                   string   `json:"role"`
	Position               string   `json:"position"`
	Department             string   `json:"department"`
	DepartmentLevel1       string   `json:"departmentLevel1"`
	DepartmentLevel2       string   `json:"departmentLevel2"`
	DepartmentLevel3       string   `json:"departmentLevel3"`
	ManagerID              string   `json:"managerId"`
	HRBPID                 string   `json:"hrbpId"`
	ResponsibleDepartments []string `json:"responsibleDepartments"`
	Status                 int      `json:"status"`
}

type Objective struct {
	ID         string      `json:"id"`
	Target     string      `json:"target"`
	Weight     float64     `json:"weight"`
	Completion interface{} `json:"completion,omitempty"`
	Result     string      `json:"result,omitempty"`
}

type NextObjective struct {
	ID     string  `json:"id"`
	Target string  `json:"target"`
	Weight float64 `json:"weight"`
}

type ValueScore struct {
	ID      string      `json:"id"`
	Self    interface{} `json:"self"`
	Manager interface{} `json:"manager"`
	HRBP    interface{} `json:"hrbp"`
	HR      interface{} `json:"hr"`
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
	ManagerID               string          `json:"managerId"`
	HRBPID                  string          `json:"hrbpId"`
	HRBPReviewerID          string          `json:"hrbpReviewerId"`
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
	History                 []HistoryDTO    `json:"history"`
}

type ReviewListResponse struct {
	List     []ReviewDTO `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type DingTalkH5Review struct {
	EmployeeID string
	ManagerID  string
	HRBPID     string
	Status     string
}

type GradeLevel struct {
	Label       string  `json:"label"`
	Grade       string  `json:"grade"`
	Coefficient float64 `json:"coefficient"`
}

type ValueRubric struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

type ValueTemplate struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Definition string        `json:"definition"`
	Rubric     []ValueRubric `json:"rubric"`
}

type TemplateDTO struct {
	ObjectiveDefaults     []NextObjective `json:"objectiveDefaults"`
	NextObjectiveDefaults []NextObjective `json:"nextObjectiveDefaults"`
	GradeLevels           []GradeLevel    `json:"gradeLevels"`
	Values                []ValueTemplate `json:"values"`
}

type LoginResponse struct {
	Token    string  `json:"token"`
	UserInfo UserDTO `json:"userInfo"`
}

type BootstrapResponse struct {
	User              UserDTO      `json:"user"`
	Menus             []AppMenuDTO `json:"menus"`
	PermissionVersion int64        `json:"permissionVersion"`
}

type WorkbenchStatsDTO struct {
	Cards []WorkbenchStatCardDTO `json:"cards"`
}

type WorkbenchStatCardDTO struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

type AppMenuDTO struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Icon          string `json:"icon"`
	PermissionKey string `json:"permissionKey"`
}

type ReviewPayload struct {
	EmployeeID             string          `json:"employeeId"`
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

type UserPayload struct {
	ID                     string      `json:"id"`
	Account                string      `json:"account"`
	Name                   string      `json:"name"`
	Password               string      `json:"password"`
	Role                   string      `json:"role"`
	Position               string      `json:"position"`
	Department             string      `json:"department"`
	DepartmentLevel1       string      `json:"departmentLevel1"`
	DepartmentLevel2       string      `json:"departmentLevel2"`
	DepartmentLevel3       string      `json:"departmentLevel3"`
	ManagerID              string      `json:"managerId"`
	HRBPID                 string      `json:"hrbpId"`
	ResponsibleDepartments interface{} `json:"responsibleDepartments"`
}

type ReviewFilters struct {
	Keyword    string
	Scope      string
	Department string
	Period     string
	NextPeriod string
	Status     string
	ManagerID  string
	HRBPID     string
	Grade      string
	Page       int
	PageSize   int
}

func userDTO(user model.DingTalkH5PerfUser) UserDTO {
	return UserDTO{
		ID:                     user.Account,
		Account:                user.Account,
		Name:                   user.Name,
		Role:                   user.Role,
		Position:               user.Position,
		Department:             user.Department,
		DepartmentLevel1:       user.DepartmentLevel1,
		DepartmentLevel2:       user.DepartmentLevel2,
		DepartmentLevel3:       user.DepartmentLevel3,
		ManagerID:              user.ManagerAccount,
		HRBPID:                 user.HRBPAccount,
		ResponsibleDepartments: decodeStringList(user.ResponsibleDepartments),
		Status:                 user.Status,
	}
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
	return result
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

func normalizeList(value interface{}) []string {
	switch typed := value.(type) {
	case []string:
		return uniqueStrings(typed)
	case []interface{}:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.ReplaceAll(toString(item), "，", ","), "、", ","), ",")))
		}
		return uniqueStrings(items)
	default:
		return uniqueStrings(strings.FieldsFunc(toString(value), func(r rune) bool {
			return r == ',' || r == '，' || r == '、' || r == ';' || r == '；' || r == '\n'
		}))
	}
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
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
