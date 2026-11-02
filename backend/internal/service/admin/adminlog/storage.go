package adminlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/support/auditlog"
	"wecheckin/backend/pkg/database"
)

// Change carries an already-redacted field difference; callers must not pass
// passwords, tokens, cookies or complete sensitive request/response bodies.
type Change struct {
	Field  string          `json:"field"`
	Before json.RawMessage `json:"before"`
	After  json.RawMessage `json:"after"`
}

// CreateContext only appends. EventID is a stable, caller-supplied deduplication
// key or is generated here; duplicate event IDs return the database error.
func CreateContext(ctx context.Context, record *model.Log) error {
	db, cancel := database.WithContext(ctx)
	defer cancel()
	return CreateWithDBContext(ctx, db, record)
}

// CreateWithDBContext lets a caller include the audit row in its transaction.
func CreateWithDBContext(ctx context.Context, db *gorm.DB, record *model.Log) error {
	if record == nil {
		return invalid("record", "不能为空")
	}
	if err := prepareRecord(record); err != nil {
		return err
	}
	if db == nil {
		return ErrUnavailable
	}
	queryContext, cancel := database.QueryContext(ctx)
	defer cancel()
	return db.WithContext(queryContext).Create(record).Error
}

func prepareRecord(row *model.Log) error {
	if row.ID != 0 {
		return ErrReadOnly
	}
	if row.Category == "" {
		row.Category = CategoryOperation
		if row.Type == 1 {
			row.Category = CategoryLogin
		}
	}
	if row.Source == "" {
		row.Source = "admin"
	}
	if row.Outcome == "" {
		row.Outcome = "unknown"
	}
	if row.SnapshotStatus == "" {
		row.SnapshotStatus = auditlog.SnapshotLegacyUnknown
	}
	if !auditlog.ValidSnapshotStatus(row.SnapshotStatus) {
		return invalid("snapshotStatus", "不支持的快照状态")
	}
	if row.ActorType == "" {
		row.ActorType = "anonymous"
		if row.AdminID > 0 {
			row.ActorType = "admin"
			if oneOf(row.Source, "client", "h5app") {
				row.ActorType = "user"
			}
		}
		if row.Source == "system" {
			row.ActorType = "system"
		}
	}
	if !oneOf(row.Category, CategoryOperation, CategoryLogin) {
		return invalid("category", "不支持的日志类型")
	}
	if !oneOf(row.Source, "admin", "client", "h5app", "system") {
		return invalid("source", "不支持的来源")
	}
	if !validOutcome(row.Outcome) {
		return invalid("outcome", "不支持的结果")
	}
	if !oneOf(row.ActorType, "admin", "user", "system", "anonymous") {
		return invalid("actorType", "不支持的操作者类型")
	}
	if (oneOf(row.ActorType, "admin", "user") && row.AdminID == 0) || (oneOf(row.ActorType, "anonymous", "system") && row.AdminID != 0) {
		return invalid("adminId", "与操作者类型不一致")
	}
	if row.AffectedCount < 0 || row.DurationMs < 0 {
		return invalid("affectedCount/durationMs", "不能为负数")
	}
	if row.AddTime < 0 || row.AddTime > maxTimestamp {
		return invalid("_createTime", "时间超出范围")
	}
	for _, item := range []struct {
		name, value string
		max         int
	}{
		{"eventId", row.EventID, 64}, {"module", row.Module, 64}, {"action", row.Action, 64}, {"reasonCode", row.ReasonCode, 64}, {"reason", row.Reason, 1000},
		{"snapshotReason", row.SnapshotReason, 80},
		{"attemptedAccount", row.AttemptedAccount, 200}, {"targetType", row.TargetType, 64}, {"targetId", row.TargetID, 128}, {"targetName", row.TargetName, 200},
		{"requestId", row.RequestID, 128}, {"method", row.Method, 16}, {"path", row.Path, 1000}, {"userAgent", row.UserAgent, 1000},
		{"relatedType", row.RelatedType, 64}, {"relatedId", row.RelatedID, 128}, {"adminName", row.AdminName, 100}, {"adminDesc", row.AdminDesc, 200}, {"LOG_ADD_IP", row.AddIP, 50},
	} {
		if err := validateText(item.name, item.value, item.max); err != nil {
			return err
		}
	}
	if len(row.Content) > 65535 {
		return invalid("content", "内容过长")
	}
	if len(row.Changes) == 0 || string(row.Changes) == "null" {
		row.Changes = json.RawMessage(`[]`)
	}
	if err := validateChanges(row.Changes); err != nil {
		return err
	}
	if row.EventID == "" {
		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			return err
		}
		row.EventID = "audit-" + hex.EncodeToString(b[:])
	}
	if row.AddTime == 0 {
		row.AddTime = database.Now()
	}
	if row.Category == CategoryLogin {
		row.Type = 1
	} else if row.Type == 1 || row.Type == 0 {
		row.Type = 2
	}
	row.UpdateBy, row.UpdateDeptID, row.EditTime = row.AdminID, row.DeptID, row.AddTime
	return nil
}

func validateChanges(raw json.RawMessage) error {
	if len(raw) > 65535 {
		return invalid("changes", "变更内容最多65535字节")
	}
	var changes []Change
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if !json.Valid(raw) || decoder.Decode(&changes) != nil || changes == nil || len(changes) > 100 {
		return invalid("changes", "必须为最多100项的字段变更数组")
	}
	for _, change := range changes {
		if strings.TrimSpace(change.Field) == "" || len(change.Before) == 0 || len(change.After) == 0 {
			return invalid("changes", "每项必须包含field、before和after")
		}
		if err := validateText("changes.field", change.Field, 128); err != nil {
			return err
		}
	}
	return nil
}

func normalizeRead(row *model.Log) {
	if row.SnapshotStatus == "" {
		row.SnapshotStatus = auditlog.SnapshotLegacyUnknown
	}
	if raw := strings.TrimSpace(string(row.TargetIDs)); raw == "" || raw == "null" {
		row.TargetIDs = json.RawMessage(`[]`)
	}
	if row.EventID == "" {
		row.EventID = fmt.Sprintf("legacy-%d", row.ID)
	}
	if row.Category == "" {
		row.Category = CategoryOperation
		if row.Type == 1 {
			row.Category = CategoryLogin
		}
	}
	if row.Source == "" {
		row.Source = "admin"
	}
	if row.ActorType == "" {
		row.ActorType = "admin"
	}
	if row.Outcome == "" {
		row.Outcome = "unknown"
	}
	if len(row.Changes) == 0 || string(row.Changes) == "null" {
		row.Changes = json.RawMessage(`[]`)
	}
}
