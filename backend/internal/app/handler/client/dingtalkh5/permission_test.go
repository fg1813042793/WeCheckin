package dingtalkh5

import "testing"

func TestDingTalkH5RoutePermissionPrefersSpecificExportRoute(t *testing.T) {
	key, ok := dingTalkH5RoutePermission("GET", "/api/v2/dingtalk/h5/reviews/export")
	if !ok || key != "dingtalk_h5:api:review:export" {
		t.Fatalf("export route mapped to %q ok=%v", key, ok)
	}
}

func TestDingTalkH5RoutePermissionMapsFlowActions(t *testing.T) {
	tests := []struct {
		path string
		key  string
	}{
		{"/api/v2/dingtalk/h5/reviews/R202607/submit-self", "dingtalk_h5:api:review:self_submit"},
		{"/api/v2/dingtalk/h5/reviews/R202607/submit-hrbp", "dingtalk_h5:api:review:hrbp_submit"},
		{"/api/v2/dingtalk/h5/reviews/R202607/finalize", "dingtalk_h5:api:review:finalize"},
	}
	for _, tt := range tests {
		key, ok := dingTalkH5RoutePermission("POST", tt.path)
		if !ok || key != tt.key {
			t.Fatalf("%s mapped to %q ok=%v, want %q", tt.path, key, ok, tt.key)
		}
	}
}
