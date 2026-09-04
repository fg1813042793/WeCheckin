package dingtalk

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestBuildBindingUserTreeKeepsDepartmentHierarchy(t *testing.T) {
	tree := buildBindingUserTree(
		[]model.Department{{ID: 1, Name: "研发"}, {ID: 2, Name: "平台", ParentID: 1}},
		[]model.User{{ID: 7, Name: "Alice", Account: "alice", Status: 1}},
		[]model.UserDept{{UserID: 7, DeptID: 2}},
	)
	if len(tree) != 1 || tree[0].Label != "研发" || tree[0].Count != 1 || len(tree[0].Children) != 1 || len(tree[0].Children[0].Children) != 1 {
		t.Fatalf("tree = %#v", tree)
	}
}

func TestUserBindingListUsesStableTopLevelFields(t *testing.T) {
	encoded, err := json.Marshal(UserBindingList{
		List: []UserBinding{}, CorpOptions: []UserBindingCorpOption{},
		UserOptions: []UserBindingUserOption{}, UserTreeOptions: []*UserBindingTreeNode{},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, field := range []string{`"list":[]`, `"total":0`, `"corpOptions":[]`, `"userOptions":[]`, `"userTreeOptions":[]`} {
		if !strings.Contains(text, field) {
			t.Fatalf("JSON %s missing %s", text, field)
		}
	}
}
