package user

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestExpandPerfUsersByDepartmentPathsDuplicatesMultiDepartmentUser(t *testing.T) {
	users := []model.DingTalkH5PerfUser{
		{ID: 7, Account: "lip", Name: "Lip", Department: "旧部门"},
	}
	paths := map[uint][]perfUserDepartmentPath{
		7: {
			{DeptID: 10, Levels: []string{"米视科技", "研发部"}},
			{DeptID: 20, Levels: []string{"米视科技", "产品部", "国内组"}},
		},
	}

	got := expandPerfUsersByDepartmentPaths(users, paths)

	if len(got) != 2 {
		t.Fatalf("expected one row per department, got %d", len(got))
	}
	if got[0].Account != "lip" || got[1].Account != "lip" {
		t.Fatalf("expected duplicated rows to keep the same account, got %#v", got)
	}
	if got[0].DepartmentID != 10 || got[0].Department != "米视科技 / 研发部" {
		t.Fatalf("unexpected first department row: %#v", got[0])
	}
	if got[1].DepartmentID != 20 || got[1].Department != "米视科技 / 产品部 / 国内组" {
		t.Fatalf("unexpected second department row: %#v", got[1])
	}
}

func TestExpandPerfUsersByDepartmentPathsKeepsUserWithoutDepartment(t *testing.T) {
	users := []model.DingTalkH5PerfUser{
		{ID: 8, Account: "amy", Name: "Amy", Department: "未设置部门"},
	}

	got := expandPerfUsersByDepartmentPaths(users, nil)

	if len(got) != 1 {
		t.Fatalf("expected user without department to remain visible, got %d", len(got))
	}
	if got[0].Account != "amy" || got[0].Department != "未设置部门" {
		t.Fatalf("unexpected user row: %#v", got[0])
	}
}

func TestExpandPerfUsersByDepartmentPathsPreservesFourthDepartmentLevel(t *testing.T) {
	users := []model.DingTalkH5PerfUser{
		{ID: 9, Account: "foster", Name: "Foster"},
	}
	paths := map[uint][]perfUserDepartmentPath{
		9: {
			{DeptID: 30, Levels: []string{"米视科技", "M/H业务", "研发部", "运维组"}},
		},
	}

	got := expandPerfUsersByDepartmentPaths(users, paths)

	if len(got) != 1 {
		t.Fatalf("expected one department row, got %d", len(got))
	}
	data, err := json.Marshal(userDTO(got[0]))
	if err != nil {
		t.Fatalf("marshal user dto: %v", err)
	}
	dtoJSON := string(data)
	if !strings.Contains(dtoJSON, `"departmentLevel3":"研发部"`) {
		t.Fatalf("expected departmentLevel3 to keep only the third level, got %s", dtoJSON)
	}
	if !strings.Contains(dtoJSON, `"departmentLevel4":"运维组"`) {
		t.Fatalf("expected departmentLevel4 to be returned, got %s", dtoJSON)
	}
	if !strings.Contains(dtoJSON, `"departmentLevels":["米视科技","M/H业务","研发部","运维组"]`) {
		t.Fatalf("expected full department levels to be returned, got %s", dtoJSON)
	}
}
