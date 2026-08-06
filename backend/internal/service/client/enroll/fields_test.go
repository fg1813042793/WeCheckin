package enroll

import "testing"

func TestDecodeEnrollObjKeepsCoverAndDescription(t *testing.T) {
	obj := decodeEnrollObj(`{"cover":["/uploads/b.png"],"desc":"打卡说明"}`)
	if len(obj.Cover) != 1 || obj.Cover[0] != "/uploads/b.png" {
		t.Fatalf("decodeEnrollObj cover = %#v, want /uploads/b.png", obj.Cover)
	}
	if obj.Desc != "打卡说明" {
		t.Fatalf("decodeEnrollObj desc = %q, want 打卡说明", obj.Desc)
	}
}
