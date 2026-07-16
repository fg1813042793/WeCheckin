package handler

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLargeHandlersUseDomainSubpackages(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "survey", "handler.go"),
		filepath.Join("admin", "survey", "formkit.go"),
		filepath.Join("admin", "survey", "formkit_types.go"),
		filepath.Join("admin", "survey", "formkit_tools.go"),
		filepath.Join("admin", "survey", "formkit_report_enroll.go"),
		filepath.Join("admin", "survey", "formkit_report_event.go"),
		filepath.Join("admin", "survey", "formkit_report_survey.go"),
		filepath.Join("admin", "survey", "formkit_csv.go"),
		filepath.Join("admin", "survey", "resource.go"),
		filepath.Join("admin", "survey", "question_bank.go"),
		filepath.Join("admin", "survey", "notification.go"),
		filepath.Join("admin", "survey", "template_presets.go"),
		filepath.Join("admin", "survey", "responses.go"),
		filepath.Join("admin", "survey", "statistics.go"),
		filepath.Join("admin", "survey", "channels.go"),
		filepath.Join("admin", "exam", "handler.go"),
		filepath.Join("admin", "exam", "resource.go"),
		filepath.Join("admin", "exam", "question_bank.go"),
		filepath.Join("admin", "exam", "records.go"),
		filepath.Join("admin", "event", "handler.go"),
		filepath.Join("admin", "event", "events.go"),
		filepath.Join("admin", "event", "participants.go"),
		filepath.Join("admin", "event", "dynamics.go"),
		filepath.Join("admin", "event", "scores.go"),
		filepath.Join("admin", "event", "helpers.go"),
		filepath.Join("admin", "enroll", "handler.go"),
		filepath.Join("admin", "enroll", "projects.go"),
		filepath.Join("admin", "enroll", "project_actions.go"),
		filepath.Join("admin", "enroll", "participants.go"),
		filepath.Join("admin", "enroll", "export.go"),
		filepath.Join("admin", "mgr", "handler.go"),
		filepath.Join("admin", "mgr", "logs.go"),
		filepath.Join("admin", "mgr", "online.go"),
		filepath.Join("admin", "user", "handler.go"),
		filepath.Join("admin", "user", "form_fields.go"),
		filepath.Join("admin", "user", "export.go"),
		filepath.Join("admin", "user", "online.go"),
		filepath.Join("admin", "news", "handler.go"),
		filepath.Join("admin", "home", "handler.go"),
		filepath.Join("admin", "setup", "handler.go"),
		filepath.Join("admin", "dict", "handler.go"),
		filepath.Join("admin", "department", "handler.go"),
		filepath.Join("admin", "role", "handler.go"),
		filepath.Join("admin", "menu", "handler.go"),
		filepath.Join("client", "survey", "handler.go"),
		filepath.Join("client", "survey", "browse.go"),
		filepath.Join("client", "survey", "logic.go"),
		filepath.Join("client", "survey", "submit.go"),
		filepath.Join("client", "survey", "responses.go"),
		filepath.Join("client", "survey", "public_tools.go"),
		filepath.Join("client", "exam", "handler.go"),
		filepath.Join("client", "exam", "browse.go"),
		filepath.Join("client", "exam", "attempt.go"),
		filepath.Join("client", "exam", "validate.go"),
		filepath.Join("client", "exam", "start.go"),
		filepath.Join("client", "exam", "save_answer.go"),
		filepath.Join("client", "exam", "submit.go"),
		filepath.Join("client", "exam", "record.go"),
		filepath.Join("client", "exam", "limits.go"),
		filepath.Join("client", "event", "handler.go"),
		filepath.Join("client", "event", "dynamics.go"),
		filepath.Join("client", "event", "scores.go"),
		filepath.Join("client", "enroll", "handler.go"),
		filepath.Join("client", "passport", "handler.go"),
		filepath.Join("client", "news", "handler.go"),
		filepath.Join("client", "favorite", "handler.go"),
		filepath.Join("public", "home", "handler.go"),
		filepath.Join("public", "geo", "handler.go"),
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("large handler implementation must live in %s: %v", file, err)
		}
	}

	for _, file := range []string{
		"survey_admin.go",
		"survey_formkit.go",
		"exam_admin.go",
		"event_admin.go",
		"enroll_admin.go",
		"admin_mgr.go",
		"user.go",
		"news_admin.go",
		"home_admin.go",
		"setup.go",
		"dict.go",
		"department.go",
		"role.go",
		"menu.go",
		"survey.go",
		"exam.go",
		"event.go",
		"enroll.go",
		"passport.go",
		"news.go",
		"fav.go",
		"home.go",
		"geo.go",
	} {
		assertNoRootHandler(t, file)
	}
}

func TestAdminExamHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "exam", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"ResourceUpload",
		"ResourceList",
		"ResourceDelete",
		"QuestionBankList",
		"QuestionBankInsert",
		"QuestionBankEdit",
		"QuestionBankDel",
		"QuestionBankCategories",
		"RecordList",
		"RecordDetail",
		"RecordDel",
		"RecordBatchDel",
		"Statistics",
	} {
		if strings.Contains(text, "func (h *AdminExamHandler) "+method) {
			t.Fatalf("%s should live outside admin/exam/handler.go", method)
		}
	}
}

func TestAdminUserHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "user", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"GetUserFormFields",
		"SaveUserFormFields",
		"UserDataGet",
		"UserDataExport",
		"UserDataDel",
		"GetOnlineUsers",
		"ForceOfflineUser",
		"BatchForceOfflineUser",
	} {
		if strings.Contains(text, "func (h *AdminUserHandler) "+method) {
			t.Fatalf("%s should live outside admin/user/handler.go", method)
		}
	}
}

func TestHighFrequencyAdminResponsesUseTypedDTOs(t *testing.T) {
	for _, file := range []string{
		filepath.Join("admin", "survey", "handler.go"),
		filepath.Join("admin", "survey", "responses.go"),
		filepath.Join("admin", "exam", "handler.go"),
		filepath.Join("admin", "exam", "records.go"),
		filepath.Join("admin", "user", "handler.go"),
		filepath.Join("admin", "event", "events.go"),
		filepath.Join("admin", "event", "participants.go"),
		filepath.Join("admin", "enroll", "projects.go"),
		filepath.Join("admin", "enroll", "participants.go"),
		filepath.Join("admin", "mgr", "logs.go"),
		filepath.Join("admin", "news", "handler.go"),
		filepath.Join("admin", "survey", "channels.go"),
		filepath.Join("admin", "survey", "notification.go"),
		filepath.Join("admin", "survey", "question_bank.go"),
		filepath.Join("admin", "survey", "statistics.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			"response.JSON(c, map[string]interface{}",
			"response.JSON(c, map[string]any",
			"survey := map[string]interface{}",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s high-frequency responses must use typed DTOs instead of %q", file, snippet)
			}
		}
	}
}

func TestClientCoreResponsesUseTypedDTOs(t *testing.T) {
	for _, file := range []string{
		filepath.Join("client", "survey", "browse.go"),
		filepath.Join("client", "survey", "logic.go"),
		filepath.Join("client", "survey", "responses.go"),
		filepath.Join("client", "survey", "submit.go"),
		filepath.Join("client", "survey", "public_tools.go"),
		filepath.Join("client", "exam", "browse.go"),
		filepath.Join("client", "exam", "start.go"),
		filepath.Join("client", "exam", "record.go"),
		filepath.Join("client", "exam", "submit.go"),
		filepath.Join("client", "exam", "validate.go"),
		filepath.Join("client", "news", "handler.go"),
		filepath.Join("client", "passport", "handler.go"),
		filepath.Join("client", "event", "handler.go"),
		filepath.Join("client", "enroll", "handler.go"),
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		text := string(src)
		for _, snippet := range []string{
			"response.JSON(c, map[string]interface{}",
			"c.JSON(consts.StatusOK, map[string]interface{}",
		} {
			if strings.Contains(text, snippet) {
				t.Fatalf("%s client core responses must use typed DTOs instead of %q", file, snippet)
			}
		}
	}
}

func TestAdminSurveyHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "survey", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"ResourceUpload",
		"ResourceList",
		"ResourceDelete",
		"QuestionBankList",
		"QuestionBankInsert",
		"QuestionBankEdit",
		"QuestionBankDel",
		"QuestionBankCategories",
		"NotifyList",
		"NotifyRead",
		"NotifyUnreadCount",
		"TemplatePresetsGet",
		"TemplatePresetsSave",
		"ResponseList",
		"ResponseDetail",
		"ResponseDel",
		"ResponseBatchDel",
		"ResponseExport",
		"Statistic",
		"ChannelList",
		"ChannelInsert",
		"ChannelDel",
	} {
		if strings.Contains(text, "func (h *AdminSurveyHandler) "+method) {
			t.Fatalf("%s should live outside admin/survey/handler.go", method)
		}
	}
}

func TestAdminSurveyFormkitKeepsToolFilesSplit(t *testing.T) {
	formkitPath := filepath.Join("admin", "survey", "formkit.go")
	src, err := os.ReadFile(formkitPath)
	if err != nil {
		t.Fatalf("read %s: %v", formkitPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"ListTypes",
		"ParseSchema",
		"EvalExpr",
		"ValidateAnswers",
		"ApplyForm",
		"ReportEnrollSchema",
		"ExportEnrollSchemaCSV",
		"ReportEventSchema",
		"ExportEventSchemaCSV",
		"ReportSurveySchema",
		"ExportSurveySchemaCSV",
	} {
		if strings.Contains(text, "func (h *AdminSurveyHandler) "+method) {
			t.Fatalf("%s should live outside admin/survey/formkit.go", method)
		}
	}
	if strings.Contains(text, "func writeCSV") {
		t.Fatalf("writeCSV should live outside admin/survey/formkit.go")
	}
}

func TestAdminEventHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "event", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"GetAdminEventList",
		"GetAdminEventDetail",
		"InsertEvent",
		"EditEvent",
		"DelEvent",
		"DelEvents",
		"StatusEvent",
		"GetEventParticipantList",
		"DelEventParticipant",
		"EditEventParticipant",
		"DelEventParticipants",
		"PostEventDynamic",
		"GetEventDynamics",
		"EditEventDynamic",
		"DelEventDynamic",
		"DelEventDynamics",
		"GetEventScores",
		"EditEventScore",
		"GetDeptUsers",
		"VouchEvent",
		"TopEvent",
	} {
		if strings.Contains(text, "func (h *AdminEventHandler) "+method) {
			t.Fatalf("%s should live outside admin/event/handler.go", method)
		}
	}
	if strings.Contains(text, "func parseUserArray") {
		t.Fatalf("parseUserArray should live outside admin/event/handler.go")
	}
}

func TestAdminEnrollHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "enroll", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"GetAdminEnrollList",
		"InsertEnroll",
		"GetEnrollDetail",
		"EditEnroll",
		"UpdateEnrollForms",
		"ClearEnrollAll",
		"DelEnroll",
		"DelEnrolls",
		"SortEnroll",
		"VouchEnroll",
		"StatusEnroll",
		"GetEnrollUserList",
		"GetEnrollStats",
		"GetEnrollJoinList",
		"RemoveEnrollUser",
		"RemoveEnrollUsers",
		"EditEnrollUserForms",
		"DelEnrollJoin",
		"DelEnrollJoins",
		"EnrollJoinDataGet",
		"EnrollJoinDataExport",
		"EnrollJoinDataDel",
	} {
		if strings.Contains(text, "func (h *AdminEnrollHandler) "+method) {
			t.Fatalf("%s should live outside admin/enroll/handler.go", method)
		}
	}
}

func TestAdminEnrollProjectsKeepActionsSplit(t *testing.T) {
	projectsPath := filepath.Join("admin", "enroll", "projects.go")
	src, err := os.ReadFile(projectsPath)
	if err != nil {
		t.Fatalf("read %s: %v", projectsPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"UpdateEnrollForms",
		"ClearEnrollAll",
		"DelEnroll",
		"DelEnrolls",
		"SortEnroll",
		"VouchEnroll",
		"StatusEnroll",
	} {
		if strings.Contains(text, "func (h *AdminEnrollHandler) "+method) {
			t.Fatalf("%s should live outside admin/enroll/projects.go", method)
		}
	}
}

func TestAdminMgrHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("admin", "mgr", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"GetLogList",
		"ClearLog",
		"AdminLogout",
		"GetOnlineAdmins",
		"ForceOfflineAdmin",
		"BatchForceOfflineAdmin",
	} {
		if strings.Contains(text, "func (h *AdminMgrHandler) "+method) {
			t.Fatalf("%s should live outside admin/mgr/handler.go", method)
		}
	}
}

func TestClientExamHandlerKeepsFlowFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("client", "exam", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"List",
		"View",
		"Validate",
		"Start",
		"SaveAnswer",
		"Submit",
		"Record",
		"MyRecords",
		"ResultBySession",
	} {
		if strings.Contains(text, "func (h *ClientExamHandler) "+method) {
			t.Fatalf("%s should live outside client/exam/handler.go", method)
		}
	}
	if strings.Contains(text, "func checkExamLimit") {
		t.Fatalf("checkExamLimit should live outside client/exam/handler.go")
	}
}

func TestClientExamAttemptKeepsMethodFilesSplit(t *testing.T) {
	attemptPath := filepath.Join("client", "exam", "attempt.go")
	src, err := os.ReadFile(attemptPath)
	if err != nil {
		t.Fatalf("read %s: %v", attemptPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"Validate",
		"Start",
		"SaveAnswer",
		"Submit",
	} {
		if strings.Contains(text, "func (h *ClientExamHandler) "+method) {
			t.Fatalf("%s should live outside client/exam/attempt.go", method)
		}
	}
}

func TestClientEventHandlerKeepsFeatureFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("client", "event", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"PostEventDynamic",
		"GetEventDynamics",
		"SaveEventScore",
		"GetEventScores",
	} {
		if strings.Contains(text, "func (h *EventHandler) "+method) {
			t.Fatalf("%s should live outside client/event/handler.go", method)
		}
	}
}

func TestClientSurveyHandlerKeepsFlowFilesSplit(t *testing.T) {
	handlerPath := filepath.Join("client", "survey", "handler.go")
	src, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatalf("read %s: %v", handlerPath, err)
	}
	text := string(src)
	for _, method := range []string{
		"List",
		"Detail",
		"ApplyLogic",
		"Validate",
		"Submit",
		"MyResponses",
		"MyResponseDetail",
		"PublicValidate",
		"PublicApply",
	} {
		if strings.Contains(text, "func (h *ClientSurveyHandler) "+method) {
			t.Fatalf("%s should live outside client/survey/handler.go", method)
		}
	}
}

func assertNoRootHandler(t *testing.T, file string) {
	t.Helper()
	if _, err := os.Stat(file); err == nil {
		t.Fatalf("%s should not remain in the root handler package; import the domain handler subpackage directly", file)
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
