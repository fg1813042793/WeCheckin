package role

import (
	"os"
	"strings"
	"testing"
)

func TestRoleWritesUseDelegationCeilingAndScopedTargets(t *testing.T) {
	src, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, snippet := range []string{
		"permissiondelegation.ResolveByAdminIDContext(ctx, tx, adminID)",
		"ceiling.ValidateRoleGrant",
		"scopedRoleQuery(ctx, tx, adminID)",
		"func DeleteForAdminContext(ctx context.Context, adminID, id uint) error",
		"func BatchDeleteForAdminContext(ctx context.Context, adminID uint, ids []uint) error",
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("role delegation enforcement missing %q", snippet)
		}
	}
}

