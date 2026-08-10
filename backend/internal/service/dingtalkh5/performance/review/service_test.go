package review

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestNextStatusAfterSelfSubmit(t *testing.T) {
	withManager := DingTalkH5Review{EmployeeID: "lip", ManagerID: "cube", HRBPID: "lucky", Status: ReviewStatusDraft}
	if got := nextStatusAfterSelfSubmit(withManager); got != ReviewStatusManagerReview {
		t.Fatalf("next status with manager = %q, want %q", got, ReviewStatusManagerReview)
	}

	withoutManager := DingTalkH5Review{EmployeeID: "nick", ManagerID: "", HRBPID: "hrbp", Status: ReviewStatusDraft}
	if got := nextStatusAfterSelfSubmit(withoutManager); got != ReviewStatusHRFinal {
		t.Fatalf("next status without manager = %q, want %q", got, ReviewStatusHRFinal)
	}
}

func TestValidateSelfSubmitPayloadRequiresAllSelfSections(t *testing.T) {
	emptyPayload := ReviewPayload{
		Objectives:     []Objective{{ID: "obj-1", Target: "  ", Weight: 40, Completion: "", Result: ""}},
		SelfSummary:    " ",
		Values:         []ValueScore{{ID: "team", Self: ""}},
		NextObjectives: []NextObjective{{ID: "next-1", Target: " ", Weight: 100}},
	}
	err := validateSelfSubmitPayload(emptyPayload)
	if err == nil {
		t.Fatalf("empty self submit payload should be rejected")
	}
	for _, want := range []string{"本月目标", "思考总结", "价值观自评", "下月目标"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validation error should mention %s, got %q", want, err.Error())
		}
	}

	completePayload := ReviewPayload{
		Objectives: []Objective{
			{ID: "obj-1", Target: "完成核心项目", Weight: 60, Completion: 100, Result: "已完成上线"},
			{ID: "obj-2", Target: "提升服务稳定性", Weight: 40, Completion: "80", Result: "完成监控优化"},
		},
		SelfSummary: "本月完成核心目标，沉淀问题复盘。",
		Values: []ValueScore{
			{ID: "team", Self: 45},
			{ID: "innovation", Self: "40"},
		},
		NextObjectives: []NextObjective{
			{ID: "next-1", Target: "推进下月版本交付", Weight: 100},
		},
	}
	if err := validateSelfSubmitPayload(completePayload); err != nil {
		t.Fatalf("complete self submit payload should pass validation: %v", err)
	}
}

func TestManagerReviewStartedDetectsManagerEvaluationFields(t *testing.T) {
	tests := []struct {
		name   string
		review model.DingTalkH5PerfReview
		want   bool
	}{
		{
			name:   "empty manager fields",
			review: model.DingTalkH5PerfReview{},
			want:   false,
		},
		{
			name:   "manager grade exists",
			review: model.DingTalkH5PerfReview{ManagerGrade: " B+ "},
			want:   true,
		},
		{
			name:   "manager comment exists",
			review: model.DingTalkH5PerfReview{ManagerComment: " 已评价 "},
			want:   true,
		},
		{
			name:   "manager value score exists",
			review: model.DingTalkH5PerfReview{ValuesJSON: encodeJSON([]ValueScore{{ID: "team", Manager: 42}})},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := managerReviewStarted(tt.review); got != tt.want {
				t.Fatalf("managerReviewStarted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithdrawContextGuardsEmployeeWithdrawAfterManagerReviewStarted(t *testing.T) {
	src, err := os.ReadFile("reviews.go")
	if err != nil {
		t.Fatalf("read reviews.go: %v", err)
	}
	body := functionBody(string(src), "func WithdrawContext")
	if body == "" {
		t.Fatalf("WithdrawContext body not found")
	}

	guardIndex := strings.Index(body, "managerReviewStarted(*review)")
	statusIndex := strings.Index(body, "review.Status = ReviewStatusDraft")
	if guardIndex < 0 {
		t.Fatalf("employee withdraw should check whether manager review has started")
	}
	if statusIndex < 0 {
		t.Fatalf("employee withdraw should still be able to return to draft before manager review starts")
	}
	if guardIndex > statusIndex {
		t.Fatalf("manager review guard must run before returning review to draft")
	}
	if !strings.Contains(body, `return fmt.Errorf("上级已评价，不能撤回")`) {
		t.Fatalf("employee withdraw should return a clear error after manager review starts")
	}
	for _, snippet := range []string{
		`reason := normalizeReviewReason(payload.ReturnReason)`,
		`return fmt.Errorf("请填写撤回原因")`,
		`action += "：" + reason`,
	} {
		if !strings.Contains(body, snippet) {
			t.Fatalf("employee withdraw should record required reason snippet %q", snippet)
		}
	}
}

func TestNormalizeUserIDKeepsStableAccountCharacters(t *testing.T) {
	if got := NormalizeUserID(" Rock.Admin_01 "); got != "rock.admin_01" {
		t.Fatalf("normalized id = %q", got)
	}
}

func TestUserPayloadDoesNotOverrideAdminPosition(t *testing.T) {
	user, err := sanitizeUserPayload(UserPayload{
		ID:               "lip",
		Name:             "Lip",
		Position:         "  Java 工程师  ",
		DepartmentLevel1: "M/H业务",
		DepartmentLevel2: "研发部",
		DepartmentLevel3: "Java开发一组",
		ManagerID:        "cube",
		HRBPID:           "lucky",
	}, nil)
	if err != nil {
		t.Fatalf("sanitize user payload: %v", err)
	}
	if got := user.Position; got != "" {
		t.Fatalf("h5 payload must not set admin position, got %q", got)
	}

	dto := userDTO(user)
	if got := dto.Position; got != "" {
		t.Fatalf("dto position should stay empty until admin user_position_id is set, got %q", got)
	}
}

func TestPerfUserMetadataIgnoresLegacyPositionInUserObj(t *testing.T) {
	user := model.DingTalkH5PerfUser{
		Account:                "lip",
		Name:                   "Lip",
		Position:               "研发经理",
		Department:             "M/H业务 / 研发部",
		DepartmentLevel1:       "M/H业务",
		DepartmentLevel2:       "研发部",
		DepartmentLevel3:       "Java开发一组",
		ManagerAccount:         "david",
		HRBPAccount:            "nick",
		ResponsibleDepartments: encodeJSON([]string{"研发部", "产品部"}),
		Obj:                    `{"theme":"blue"}`,
	}

	raw := encodePerfUserObj(user.Obj, user)
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		t.Fatalf("decode user obj: %v", err)
	}
	if obj["theme"] != "blue" {
		t.Fatalf("existing user_obj keys should be preserved, got %#v", obj["theme"])
	}
	if strings.Contains(raw, `"role"`) {
		t.Fatalf("performance role must not be written into user_obj, got %s", raw)
	}
	if strings.Contains(raw, `"position"`) {
		t.Fatalf("admin position must not be written into dingtalk h5 user_obj, got %s", raw)
	}

	stored := model.DingTalkH5PerfUser{Account: "lip", Name: "Lip", Obj: raw, Status: 1}
	hydratePerfUser(&stored)
	if stored.Role != "" || stored.Position != "" || stored.ManagerAccount != "david" || stored.HRBPAccount != "nick" {
		t.Fatalf("metadata was not restored: %#v", stored)
	}
	if got := decodeStringList(stored.ResponsibleDepartments); len(got) != 2 || got[0] != "研发部" || got[1] != "产品部" {
		t.Fatalf("responsible departments = %#v", got)
	}
}

func TestPerfUserLegacyPositionIsIgnoredWhenAdminPositionMissing(t *testing.T) {
	stored := model.DingTalkH5PerfUser{
		Account: "cube",
		Name:    "Cube",
		Obj:     `{"dingtalkH5Performance":{"position":"产品主管","managerId":"nick","hrbpId":"lucky"}}`,
		Status:  1,
	}
	hydratePerfUser(&stored)
	if got := stored.Position; got != "" {
		t.Fatalf("legacy h5 position should be ignored without user_position_id, got %q", got)
	}
	if stored.ManagerAccount != "nick" || stored.HRBPAccount != "lucky" {
		t.Fatalf("manager/hrbp metadata should still be restored: %#v", stored)
	}
}

func TestCurrentObjectivesPreferPreviousNextObjectives(t *testing.T) {
	previous := &model.DingTalkH5PerfReview{
		ReviewNo:           "lip-2026-07",
		Period:             "2026-07",
		NextPeriod:         "2026-08",
		NextObjectivesJSON: encodeJSON([]NextObjective{{ID: "old-next-1", Target: "继续优化移动端体验", Weight: 60}, {ID: "old-next-2", Target: "补齐核心链路监控", Weight: 40}}),
	}
	templateDefaults := []NextObjective{{ID: "tpl-1", Target: "模板默认目标", Weight: 100}}

	objectives, source := currentObjectivesForNewReview("lip-2026-08", templateDefaults, previous)

	if source.reviewNo != "lip-2026-07" || source.period != "2026-07" {
		t.Fatalf("objective source = %#v, want previous review source", source)
	}
	if len(objectives) != 2 {
		t.Fatalf("objectives = %d, want 2: %#v", len(objectives), objectives)
	}
	if objectives[0].ID != "lip-2026-08-obj-1" || objectives[0].Target != "继续优化移动端体验" || objectives[0].Weight != 60 {
		t.Fatalf("first inherited objective = %#v", objectives[0])
	}
	if objectives[1].ID != "lip-2026-08-obj-2" || objectives[1].Target != "补齐核心链路监控" || objectives[1].Weight != 40 {
		t.Fatalf("second inherited objective = %#v", objectives[1])
	}
	if objectives[0].Completion != "" || objectives[0].Result != "" {
		t.Fatalf("inherited objective should reset self-evaluation fields, got %#v", objectives[0])
	}
}

func TestCurrentObjectivesFallbackToTemplateWhenPreviousNextObjectivesEmpty(t *testing.T) {
	previous := &model.DingTalkH5PerfReview{
		ReviewNo:           "lip-2026-07",
		Period:             "2026-07",
		NextPeriod:         "2026-08",
		NextObjectivesJSON: encodeJSON([]NextObjective{{ID: "empty", Target: "  ", Weight: 100}}),
	}

	objectives, source := currentObjectivesForNewReview("lip-2026-08", []NextObjective{{ID: "tpl-1", Target: "模板默认目标", Weight: 100}}, previous)

	if source.reviewNo != "" || source.period != "" {
		t.Fatalf("empty previous next objectives should not set source, got %#v", source)
	}
	if len(objectives) != 1 || objectives[0].Target != "模板默认目标" || objectives[0].ID != "lip-2026-08-obj-1" {
		t.Fatalf("objectives should fall back to template defaults, got %#v", objectives)
	}
}
