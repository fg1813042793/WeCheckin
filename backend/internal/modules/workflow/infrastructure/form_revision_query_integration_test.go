package infrastructure

import (
	"strings"
	"testing"

	"wecheckin/backend/internal/modules/workflow/application"
)

func TestUnifiedTaskQueryIncludesPendingFormRevisionTasks(t *testing.T) {
	sqlText, args := unifiedTaskListSQL(application.TaskQuery{
		AssigneeID: "42", Status: "pending", InstanceTitle: "采购",
		DefinitionName: "采购审批", DefinitionCategory: "purchase",
		StarterName: "张", StartTimeFrom: 1000, StartTimeTo: 2000,
		Page: 2, PageSize: 20, IncludeFormRevisions: true,
	})
	normalized := strings.Join(strings.Fields(sqlText), " ")
	for _, fragment := range []string{
		"FROM workflow_process_tasks workflow_task",
		"UNION ALL",
		"FROM workflow_form_revision_tasks revision_task",
		"revision_task.revision_request_id AS revision_request_id",
		"'form_revision' AS task_type",
		"ORDER BY unified_task.sort_time DESC, unified_task.id DESC",
		"LIMIT ? OFFSET ?",
		"instance_title LIKE ? ESCAPE '!'",
		"definition_category = ?",
	} {
		if !strings.Contains(normalized, fragment) {
			t.Fatalf("unified task query missing %q: %s", fragment, normalized)
		}
	}
	if len(args) < 2 || args[len(args)-2] != 20 || args[len(args)-1] != 20 {
		t.Fatalf("pagination args=%#v", args)
	}
}

func TestRevisionTaskRowMapsTaskDiscriminator(t *testing.T) {
	summary, err := unifiedTaskSummary(unifiedTaskRow{
		ID: "revision-task-1", InstanceID: "instance-1", NodeID: "finance", NodeName: "财务确认",
		AssigneeID: "42", Status: "pending", TaskType: "form_revision", RevisionRequestID: "revision-1",
		ImagesJSON: `[]`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.TaskType != "form_revision" || summary.RevisionRequestID != "revision-1" || summary.InstanceID != "instance-1" {
		t.Fatalf("summary=%#v", summary)
	}
}
