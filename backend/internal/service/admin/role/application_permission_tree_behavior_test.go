package role

import (
	"os"
	"strings"
	"testing"
)

func TestApplicationPermissionTreeContainsBuiltInMenus(t *testing.T) {
	tree := ApplicationPermissionTree()
	if len(tree.Client) == 0 {
		t.Fatalf("client application permission tree must not be empty")
	}
	if len(tree.DingTalkH5) == 0 {
		t.Fatalf("dingtalk h5 application permission tree must not be empty")
	}
	var performance *ApplicationPermissionNode
	for i := range tree.DingTalkH5 {
		if tree.DingTalkH5[i].Key == "dingtalk_h5:menu:performance" {
			performance = &tree.DingTalkH5[i]
			break
		}
	}
	if performance == nil {
		t.Fatalf("dingtalk h5 tree must contain performance root")
	}
	if len(performance.Children) == 0 {
		t.Fatalf("performance root must contain tab children")
	}
	if findApplicationPermissionNode(performance.Children, "dingtalk_h5:button:review:create") == nil {
		t.Fatalf("dingtalk h5 performance tree must expose review create button permission")
	}
	org := findApplicationPermissionNode(performance.Children, "dingtalk_h5:menu:performance:org")
	if org == nil {
		t.Fatalf("dingtalk h5 performance tree must contain org menu")
	}
	if findApplicationPermissionNode(org.Children, "dingtalk_h5:button:user:config") == nil {
		t.Fatalf("dingtalk h5 org menu must expose user flow config button permission")
	}
	workflow := findApplicationPermissionNode(tree.DingTalkH5, "dingtalk_h5:menu:workflow")
	if workflow == nil {
		t.Fatalf("dingtalk h5 tree must contain workflow menu")
	}
	if findApplicationPermissionNode(workflow.Children, "dingtalk_h5:button:workflow:form-revise") == nil {
		t.Fatalf("dingtalk h5 workflow menu must expose form revise button permission")
	}
}

func findApplicationPermissionNode(nodes []ApplicationPermissionNode, key string) *ApplicationPermissionNode {
	for i := range nodes {
		if nodes[i].Key == key {
			return &nodes[i]
		}
		if found := findApplicationPermissionNode(nodes[i].Children, key); found != nil {
			return found
		}
	}
	return nil
}

func TestApplicationPermissionTreeEndpointUsesDatabaseNames(t *testing.T) {
	serviceSrc, err := os.ReadFile("service.go")
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}
	handlerSrc, err := os.ReadFile("../../../handler/admin/role/handler.go")
	if err != nil {
		t.Fatalf("read handler.go: %v", err)
	}
	combined := string(serviceSrc) + string(handlerSrc)
	for _, snippet := range []string{
		"func ApplicationPermissionTreeContext(ctx context.Context) ApplicationPermissionTreeResponse",
		"applicationPermissionLabelsContext(ctx, db, permissionsupport.PlatformDingTalkH5, permissionsupport.TypeDirectory, permissionsupport.TypeMenu, permissionsupport.TypeButton)",
		"appmenuperm.DingTalkH5PermissionDeclarations()",
		"roleservice.ApplicationPermissionTreeContext(ctx)",
	} {
		if !strings.Contains(combined, snippet) {
			t.Fatalf("application permission tree must use current permissions table labels with %q", snippet)
		}
	}
}

func TestDingTalkH5BuiltInParentSupportsDirectoryType(t *testing.T) {
	src, err := os.ReadFile("../../../support/appmenuperm/catalog.go")
	if err != nil {
		t.Fatalf("read app menu catalog: %v", err)
	}
	text := string(src)
	for _, snippet := range []string{
		`TypeDirectory = "directory"`,
		`Key: "dingtalk_h5:menu:performance", Name: "绩效管理", Platform: "dingtalk_h5", Type: TypeDirectory`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("dingtalk h5 built-in parent menu must support directory declaration with %q", snippet)
		}
	}
}
