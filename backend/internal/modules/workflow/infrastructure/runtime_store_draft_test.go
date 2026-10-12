package infrastructure

import (
	"testing"
	"time"

	"gorm.io/gorm"

	workflowmodel "wecheckin/backend/internal/model/workflow"
)

func TestStartDraftUpsertUsesDatetimeForUpdatedAt(t *testing.T) {
	db := openWorkflowSearchDryRunDB(t).Session(&gorm.Session{SkipDefaultTransaction: true})
	updatedAtMillis := int64(1790668445906)
	updatedAt := time.UnixMilli(updatedAtMillis)
	row := workflowmodel.StartDraft{
		DefinitionID: 1, DefinitionVersion: 23, StarterID: "42",
		FormDataJSON: `{"field":"updated"}`, EditTime: updatedAtMillis,
		UpdatedAt: updatedAt,
	}

	statement := upsertStartDraftRow(db, &row).Statement
	matchingTimes := 0
	for _, variable := range statement.Vars {
		value, ok := variable.(time.Time)
		if ok && value.Equal(updatedAt) {
			matchingTimes++
		}
	}
	if matchingTimes < 2 {
		t.Fatalf("start draft upsert vars = %#v, want datetime for insert and conflict update", statement.Vars)
	}
}
