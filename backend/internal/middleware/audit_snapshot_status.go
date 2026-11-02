package middleware

import (
	"encoding/json"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
)

// Snapshot availability is independent of whether the business operation succeeded.
func finalizeAuditSnapshot(event *model.Log) {
	if event.Outcome == "failure" || event.Outcome == "blocked" {
		event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotOperationFailed, "business_not_successful"
		event.Changes = auditlog.SafeChanges(nil)
		return
	}
	if event.Outcome == "unknown" {
		event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotUnavailable, "result_unconfirmed"
		return
	}
	if event.SnapshotStatus == auditlog.SnapshotUnavailable || event.SnapshotStatus == auditlog.SnapshotNotConfigured {
		return
	}
	var changes []auditlog.Change
	if json.Unmarshal(event.Changes, &changes) == nil && len(changes) > 0 {
		event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotRecorded, ""
		return
	}
	if event.SnapshotStatus == auditlog.SnapshotUnchanged {
		event.SnapshotReason = "no_collected_changes"
		return
	}
	event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotNotApplicable, "no_field_snapshot"
	if event.Category == "login" || event.Method == "GET" {
		return
	}
	switch event.Action {
	case "password_change", "force_offline", "send", "test_send", "run", "retry", "export", "bind", "logout":
		return
	case "create":
		event.SnapshotReason = "create_target_only"
		return
	}
	switch event.TargetType {
	case "workflow_instance", "workflow_task", "form_revision", "form_revision_task", "scheduled_task_run":
		event.SnapshotReason = "runtime_history"
		return
	}
	event.SnapshotStatus, event.SnapshotReason = auditlog.SnapshotNotConfigured, "unsupported_route"
}
