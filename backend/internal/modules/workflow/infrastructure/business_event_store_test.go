package infrastructure

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/modules/workflow/application"
)

func TestBusinessEventModelCarriesFormRevisionPayload(t *testing.T) {
	event := application.WorkflowBusinessEvent{
		ID: "event-1", DedupeKey: "workflow.instance.form_revised:revision-1",
		Event: application.LifecycleEvent{
			Type: application.LifecycleInstanceFormRevised, InstanceID: "instance-1",
			BusinessType: "purchase", BusinessKey: "PO-1", Status: "completed",
			RevisionRequestID: "revision-1", FormRevision: 4,
			FormPatch: map[string]interface{}{"remark": "x"},
		},
	}

	row, err := businessEventModel(event, 1234)
	if err != nil {
		t.Fatal(err)
	}
	if row.ID != "event-1" || row.EventType != string(application.LifecycleInstanceFormRevised) ||
		row.AggregateID != "instance-1" || row.BusinessType != "purchase" || row.BusinessKey != "PO-1" ||
		row.Status != workflowmodel.BusinessEventStatusPending || row.NextRetryAt != 1234 {
		t.Fatalf("row=%#v", row)
	}
	record, err := businessEventRecordFromModel(row)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(record.Event.FormPatch, event.Event.FormPatch) || record.Event.RevisionRequestID != "revision-1" {
		t.Fatalf("record=%#v", record)
	}
}

func TestClaimableBusinessEventQueryIncludesLeaseRecovery(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &sql.DB{}, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	var rows []workflowmodel.BusinessEventOutbox
	statement := claimableBusinessEventQuery(db, 2000, 1000).
		Limit(25).Find(&rows).Statement
	sqlText := strings.Join(strings.Fields(statement.SQL.String()), " ")
	for _, fragment := range []string{"event_status = ?", "next_retry_at <= ?", "edit_time <= ?", "ORDER BY next_retry_at ASC", "LIMIT ?", "FOR UPDATE"} {
		if !strings.Contains(sqlText, fragment) {
			t.Fatalf("claim query missing %q: %s", fragment, sqlText)
		}
	}
}
