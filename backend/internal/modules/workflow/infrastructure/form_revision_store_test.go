package infrastructure

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	workflowdomain "wecheckin/backend/internal/modules/workflow/domain"
	"wecheckin/backend/internal/workflowcore"
)

func TestFormRevisionStoreLoadsInstanceAndActiveRequestForUpdate(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t)
	instanceSQL := lockCompletedRevisionSourceQuery(db, "instance-1").Statement.SQL.String()
	requestSQL := lockActiveFormRevisionQuery(db, "instance-1").Statement.SQL.String()
	for name, sqlText := range map[string]string{"instance": instanceSQL, "request": requestSQL} {
		if !strings.Contains(strings.ToUpper(sqlText), "FOR UPDATE") {
			t.Fatalf("%s lock query=%s", name, sqlText)
		}
	}
}

func TestFormRevisionRowsRoundTrip(t *testing.T) {
	want := completedRevisionFixture()
	requestRow, taskRows, err := formRevisionModels(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := formRevisionFromModels(requestRow, taskRows)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%#v\nwant=%#v\nerr=%v", got, want, err)
	}
}

func completedRevisionFixture() *workflowdomain.FormRevisionRequest {
	return &workflowdomain.FormRevisionRequest{
		ID: "revision-1", SourceInstanceID: "instance-1", SourceNodeID: "manager", SourceNodeName: "主管审批",
		RequesterID: "42", BaseFormRevision: 3, AppliedFormRevision: 0,
		BeforeFormData:     map[string]interface{}{"remark": "原说明", "amount": json.Number("100")},
		Patch:              map[string]interface{}{"remark": "新说明"},
		ProposedFormData:   map[string]interface{}{"remark": "新说明", "amount": json.Number("100")},
		ChangedFieldLabels: []string{"说明"}, Reason: "修正录入错误",
		Mode: workflowdomain.FormRevisionModeDownstreamReview, Status: workflowdomain.FormRevisionStatusPending,
		CreatedAt: 1000,
		Tasks: []workflowdomain.FormRevisionTask{{
			ID: "revision-task-1", RevisionRequestID: "revision-1", SourceTaskID: "task-2",
			NodeID: "finance", NodeName: "财务审批", Stage: 1,
			AssigneeID: "84", AssigneeName: "Nick", ApprovalMode: workflowcore.ApprovalModeParallel,
			CompletionRate: 100, Sequence: 1, Total: 1, Status: workflowdomain.RevisionTaskStatusApproved,
			Action: workflowdomain.RevisionActionApprove, Comment: "同意",
			Images:    []workflowcore.FormAttachment{{ID: "image-1", Name: "proof.png", URL: "/uploads/proof.png", MimeType: "image/png", Size: 1024}},
			HandledBy: "84", HandledAt: 2000,
		}},
	}
}
