package publish

import "testing"

func TestHasDeptAccessAllowsEmptyPublishScope(t *testing.T) {
	if !HasDeptAccess("", []uint{1}) {
		t.Fatal("empty publish scope should be public")
	}
}

func TestHasDeptAccessMatchesUserDepartment(t *testing.T) {
	if !HasDeptAccess("1, 2,3", []uint{9, 2}) {
		t.Fatal("expected department 2 to match")
	}
	if HasDeptAccess("1,3", []uint{2}) {
		t.Fatal("unexpected department access")
	}
}

func TestDeptOverlapBuildsFindInSetExpression(t *testing.T) {
	got := DeptOverlap("news_publish_dept_ids", []uint{1, 2})
	want := "FIND_IN_SET('1', `news_publish_dept_ids`) OR FIND_IN_SET('2', `news_publish_dept_ids`)"
	if got != want {
		t.Fatalf("DeptOverlap() = %q, want %q", got, want)
	}
	if got := DeptOverlap("news_publish_dept_ids", nil); got != "1 = 0" {
		t.Fatalf("empty DeptOverlap() = %q", got)
	}
}

func TestJoinStatusDesc(t *testing.T) {
	cases := map[int]string{
		0:  "待审核",
		1:  "已通过",
		2:  "未通过",
		99: "未知",
	}
	for in, want := range cases {
		if got := JoinStatusDesc(in); got != want {
			t.Fatalf("JoinStatusDesc(%d) = %q, want %q", in, got, want)
		}
	}
}
