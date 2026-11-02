package adminpermission

import "testing"

func TestNormalizePermissionRequestLimitsHiddenToAdminMenus(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		ptype    string
		want     int
	}{
		{name: "admin directory", platform: "admin", ptype: "directory", want: 1},
		{name: "admin menu", platform: "admin", ptype: "menu", want: 1},
		{name: "admin button", platform: "admin", ptype: "button", want: 0},
		{name: "client menu", platform: "client", ptype: "menu", want: 0},
		{name: "dingtalk menu", platform: "dingtalk_h5", ptype: "menu", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := normalizePermissionRequest(SaveRequest{Key: "test:key", Name: "Test", Platform: tt.platform, Type: tt.ptype, Hidden: 1})
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			if got := item.HiddenValue(); got != tt.want {
				t.Fatalf("hidden = %d, want %d", got, tt.want)
			}
		})
	}
}
