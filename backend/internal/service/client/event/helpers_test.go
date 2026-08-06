package event

import (
	"testing"
	"time"

	"wecheckin/backend/internal/model"
)

func TestDecodeEventObjKeepsCoverDescriptionAndRules(t *testing.T) {
	obj := decodeEventObj(`{"cover":["/uploads/event.png"],"desc":"活动说明","rules":"规则"}`)
	if len(obj.Cover) != 1 || obj.Cover[0] != "/uploads/event.png" {
		t.Fatalf("decodeEventObj cover = %#v, want /uploads/event.png", obj.Cover)
	}
	if obj.Desc != "活动说明" || obj.Rules != "规则" {
		t.Fatalf("decodeEventObj = %#v, want desc/rules", obj)
	}
}

func TestPopulateEventTimeFieldsMarksDisabledEvent(t *testing.T) {
	e := &model.Event{
		Status:     0,
		RegStart:   time.Date(2026, 7, 1, 9, 0, 0, 0, time.Local).UnixMilli(),
		RegEnd:     time.Date(2026, 7, 2, 18, 0, 0, 0, time.Local).UnixMilli(),
		EventStart: time.Date(2026, 7, 3, 9, 0, 0, 0, time.Local).UnixMilli(),
		EventEnd:   time.Date(2026, 7, 4, 18, 0, 0, 0, time.Local).UnixMilli(),
	}
	populateEventTimeFields(e)
	if e.StatusDesc != "已停用" {
		t.Fatalf("StatusDesc = %q, want 已停用", e.StatusDesc)
	}
	if e.RegStartStr != "2026-07-01 09:00" || e.EventEndStr != "2026-07-04 18:00" {
		t.Fatalf("time fields not formatted: %#v", e)
	}
}
