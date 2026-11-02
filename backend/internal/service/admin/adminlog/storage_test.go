package adminlog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestLegacyNormalizationNeverClaimsSuccess(t *testing.T) {
	for _, typ := range []int{0, 1, 2} {
		row := model.Log{ID: 42, Type: typ, AdminID: 7, AdminName: "snapshot"}
		normalizeRead(&row)
		category := CategoryOperation
		if typ == 1 {
			category = CategoryLogin
		}
		if row.EventID != "legacy-42" || row.Category != category || row.Outcome != "unknown" || row.SnapshotStatus != "legacy_unknown" || row.Source != "admin" || row.ActorType != "admin" || string(row.Changes) != "[]" || row.AdminName != "snapshot" {
			t.Fatalf("legacy row: %+v", row)
		}
	}
}

func TestNormalizeReadTargetIDsPreservesRequestedAndAffectedCounts(t *testing.T) {
	for _, raw := range []string{"", "null", " \nnull\t", `["41","42","43"]`} {
		row := model.Log{ID: 41, TargetIDs: json.RawMessage(raw), RequestedCount: 3, AffectedCount: 1}
		normalizeRead(&row)
		want := "[]"
		if raw == `["41","42","43"]` {
			want = raw
		}
		if string(row.TargetIDs) != want || row.RequestedCount != 3 || row.AffectedCount != 1 {
			t.Fatalf("raw=%q normalized=%+v", raw, row)
		}
	}
}

func TestAuditStorageKeepsDefaultTimeoutAndCancellation(t *testing.T) {
	f := &auditDBFixture{}
	openAuditTestDB(t, f)
	if err := CreateContext(context.Background(), &model.Log{}); err != nil {
		t.Fatal(err)
	}
	if !f.writeHasDeadline {
		t.Error("storage dropped database default timeout")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := CreateContext(ctx, &model.Log{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled write: %v", err)
	}
	if len(f.writes) != 1 {
		t.Fatal("cancelled request wrote a record")
	}
}

func TestAuditStorageInfersUserActorForClientSources(t *testing.T) {
	for _, source := range []string{"client", "h5app"} {
		row := model.Log{Source: source, AdminID: 7}
		if err := prepareRecord(&row); err != nil {
			t.Fatal(err)
		}
		if row.ActorType != "user" {
			t.Fatalf("%s actor type: %s", source, row.ActorType)
		}
	}
}

func TestPrepareAuditRecord(t *testing.T) {
	row := model.Log{Category: CategoryLogin, Source: "h5app", ActorType: "anonymous", Outcome: "failure", AttemptedAccount: "test", Changes: json.RawMessage(`[{"field":"status","before":null,"after":1}]`)}
	if err := prepareRecord(&row); err != nil {
		t.Fatal(err)
	}
	if row.EventID == "" || row.Type != 1 || row.AddTime <= 0 || row.Outcome != "failure" {
		t.Fatalf("record: %+v", row)
	}
	second := model.Log{}
	if err := prepareRecord(&second); err != nil {
		t.Fatal(err)
	}
	if second.EventID == row.EventID || second.Outcome != "unknown" {
		t.Fatalf("record: %+v", second)
	}
}

func TestPrepareAuditRejectsInvalidRecord(t *testing.T) {
	for _, row := range []model.Log{
		{ID: 1}, {Category: "anything"}, {Source: "browser"}, {Outcome: "ok"}, {ActorType: "visitor"},
		{AffectedCount: -1}, {DurationMs: -1}, {ActorType: "anonymous", AdminID: 1}, {ActorType: "user"},
		{Changes: json.RawMessage(`{}`)}, {Changes: json.RawMessage(`[null]`)},
		{Changes: json.RawMessage(`[{"field":"x"}]`)}, {Changes: json.RawMessage(`[{"field":"","before":0,"after":1}]`)},
	} {
		if err := prepareRecord(&row); err == nil {
			t.Fatalf("accepted invalid record: %+v", row)
		}
	}
}
