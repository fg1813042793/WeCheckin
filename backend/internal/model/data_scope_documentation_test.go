package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoleDataScopeDocumentationIncludesCustomAndExtra(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	files := map[string][]string{
		filepath.Join("organization", "rbac.go"): {
			"数据范围:1全部 2本部门及子部门 3本人 4自定义部门",
		},
		filepath.Join("..", "handler", "admin", "role", "handler.go"): {
			"数据权限范围(1=全部 2=本部门及子部门 3=本人 4=自定义部门)",
		},
		filepath.Join(root, "docs", "ACCESS_CONTROL_RUNTIME.md"): {
			"`data:custom`",
			"`data:extra`",
			"角色表 `roles.role_data_scope` 仅保留 1/2/3/4 的基础范围兼容值",
			"用户级 `data:extra` 会在基础范围之外追加可见部门或可见用户",
		},
		filepath.Join(root, "docs", "DINGTALK_H5_PERFORMANCE.md"): {
			"`data:all`、`data:dept`、`data:self`、`data:custom`、`data:extra`",
		},
		filepath.Join(root, "docs", "superpowers", "specs", "2026-07-28-unified-permissions-design.md"): {
			"`data:all`、`data:dept`、`data:self`、`data:custom`、`data:extra`",
			"创建内置权限：`admin:login`、`data:all`、`data:dept`、`data:self`、`data:custom`、`data:extra`",
		},
	}
	for path, snippets := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(src)
		for _, snippet := range snippets {
			if !strings.Contains(text, snippet) {
				t.Fatalf("%s must document current data scope capability with %q", path, snippet)
			}
		}
	}

	accessDoc, err := os.ReadFile(filepath.Join(root, "docs", "ACCESS_CONTROL_RUNTIME.md"))
	if err != nil {
		t.Fatalf("read access control doc: %v", err)
	}
	if strings.Contains(string(accessDoc), "仍可能存在旧数据范围表述") {
		t.Fatalf("access control doc must not keep obsolete data scope documentation risk note")
	}
}
