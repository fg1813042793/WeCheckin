package admincontent

import "testing"

func TestDecodeEnrollObjKeepsCoverAndDescription(t *testing.T) {
	obj := decodeEnrollObj(`{"cover":["/uploads/a.png"],"desc":"活动说明"}`)
	if len(obj.Cover) != 1 || obj.Cover[0] != "/uploads/a.png" {
		t.Fatalf("decodeEnrollObj cover = %#v, want /uploads/a.png", obj.Cover)
	}
	if obj.Desc != "活动说明" {
		t.Fatalf("decodeEnrollObj desc = %q, want 活动说明", obj.Desc)
	}
}

func TestDecodeEnrollObjToleratesInvalidJSON(t *testing.T) {
	obj := decodeEnrollObj(`{broken`)
	if len(obj.Cover) != 0 || obj.Desc != "" {
		t.Fatalf("invalid JSON should return zero enrollObj, got %#v", obj)
	}
}
