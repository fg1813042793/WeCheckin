package adminlog

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestParseQueryValidation(t *testing.T) {
	for _, raw := range []string{
		"page=0", "page=-1", "page=x", "page=1000001", "pageSize=0", "pageSize=101",
		"source=unknown", "outcome=ok", "actorId=-1", "actorId=18446744073709551616",
		"ip=not-an-ip", "startTime=bad", "startTime=-1", "startTime=2000&endTime=1000",
		"endTime=253402300800000", "module=" + strings.Repeat("a", 65),
		"search=" + strings.Repeat("a", 201), "category=login", "page=1&page=2",
	} {
		t.Run(raw, func(t *testing.T) {
			values, _ := url.ParseQuery(raw)
			_, err := ParseQuery(values, false)
			var invalid *ValidationError
			if !errors.As(err, &invalid) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
	q, err := ParseQuery(url.Values{}, false)
	if err != nil || q.Page != 1 || q.PageSize != 20 {
		t.Fatalf("defaults: %+v, %v", q, err)
	}
	q, err = ParseQuery(url.Values{"actorId": {"0"}, "source": {"h5app"}, "outcome": {"unknown"}, "startTime": {"2026-10-28"}, "endTime": {"2026-10-28"}}, false)
	if err != nil || q.ActorID == nil || *q.ActorID != 0 || q.EndTime-q.StartTime != 86400000-1 {
		t.Fatalf("date/anonymous query: %+v, %v", q, err)
	}
	q, err = ParseQuery(url.Values{"startTime": {"2026-10-28T00:00:00+08:00"}}, false)
	if err != nil || q.StartTime != 1793116800000 {
		t.Fatalf("RFC3339: %+v, %v", q, err)
	}
}

func TestExportQueryBounds(t *testing.T) {
	q, err := ParseQuery(url.Values{}, true)
	if err != nil || q.PageSize != 1000 {
		t.Fatalf("export defaults: %+v %v", q, err)
	}
	if _, err := ParseQuery(url.Values{"pageSize": {"1001"}}, true); err == nil {
		t.Fatal("unbounded export accepted")
	}
}

func TestAuditLogsAreReadOnly(t *testing.T) {
	for name, f := range map[string]func() error{
		"clear": Clear, "clear context": func() error { return ClearContext(context.Background()) },
		"delete":         func() error { return DeleteIDs([]uint{1}) },
		"delete context": func() error { return DeleteIDsContext(context.Background(), nil) },
	} {
		t.Run(name, func(t *testing.T) {
			if err := f(); !errors.Is(err, ErrReadOnly) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestAuditLogModelContract(t *testing.T) {
	typ := reflect.TypeOf(model.Log{})
	for field, contract := range map[string][2]string{
		"EventID": {"eventId", "log_event_id"}, "Category": {"category", "log_category"},
		"Source": {"source", "log_source"}, "Module": {"module", "log_module"}, "Action": {"action", "log_action"},
		"Outcome": {"outcome", "log_outcome"}, "ReasonCode": {"reasonCode", "log_reason_code"}, "Reason": {"reason", "log_reason"},
		"ActorType": {"actorType", "log_actor_type"}, "AttemptedAccount": {"attemptedAccount", "log_attempted_account"},
		"TargetType": {"targetType", "log_target_type"}, "TargetID": {"targetId", "log_target_id"}, "TargetName": {"targetName", "log_target_name"},
		"AffectedCount": {"affectedCount", "log_affected_count"}, "RequestID": {"requestId", "log_request_id"},
		"Method": {"method", "log_method"}, "Path": {"path", "log_path"}, "DurationMs": {"durationMs", "log_duration_ms"},
		"UserAgent": {"userAgent", "log_user_agent"}, "Changes": {"changes", "log_changes"},
		"SnapshotStatus": {"snapshotStatus", "log_snapshot_status"}, "SnapshotReason": {"snapshotReason", "log_snapshot_reason"},
		"RelatedType": {"relatedType", "log_related_type"}, "RelatedID": {"relatedId", "log_related_id"},
		"AdminID": {"adminId", "create_by"}, "AdminName": {"adminName", "log_admin_name"}, "AdminDesc": {"adminDesc", "log_admin_desc"},
		"AddIP": {"LOG_ADD_IP", "log_add_ip"}, "AddTime": {"_createTime", "add_time"},
	} {
		f, ok := typ.FieldByName(field)
		if !ok || f.Tag.Get("json") != contract[0] || !strings.Contains(f.Tag.Get("gorm"), "column:"+contract[1]) {
			t.Errorf("%s contract mismatch: %+v", field, f)
		}
	}
	f, _ := typ.FieldByName("Changes")
	if f.Type != reflect.TypeOf(json.RawMessage{}) {
		t.Error("changes must be raw JSON")
	}
}
