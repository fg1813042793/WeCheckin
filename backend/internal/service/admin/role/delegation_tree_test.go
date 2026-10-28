package role

import "testing"

func TestFilterApplicationPermissionTreeKeepsRequiredAncestors(t *testing.T) {
	nodes := []ApplicationPermissionNode{{
		Key: "client:menu:workflow", Name: "流程", Children: []ApplicationPermissionNode{
			{Key: "client:menu:workflow:list", Name: "流程列表"},
			{Key: "client:menu:workflow:admin", Name: "流程管理"},
		},
	}}
	filtered := filterApplicationPermissionNodes(nodes, map[string]struct{}{"client:menu:workflow:list": {}})
	if len(filtered) != 1 || len(filtered[0].Children) != 1 || filtered[0].Children[0].Key != "client:menu:workflow:list" {
		t.Fatalf("filtered tree = %#v", filtered)
	}
}

