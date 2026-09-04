package adminrouteperm

import "testing"

func TestContentActionPermissionCatalogUsesButtonGranularity(t *testing.T) {
	want := map[string]bool{
		"enroll:status": false,
		"enroll:vouch":  false,
		"enroll:export": false,
		"enroll:users":  false,
		"news:status":   false,
		"news:vouch":    false,
		"event:status":  false,
		"event:vouch":   false,
		"event:top":     false,
		"event:users":   false,
		"upload:create": false,
	}
	for _, item := range Declarations() {
		if _, ok := want[item.Perms]; ok {
			want[item.Perms] = true
		}
	}
	for perms, found := range want {
		if !found {
			t.Fatalf("admin route permission catalog missing %s", perms)
		}
	}
}
