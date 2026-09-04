package infrastructure

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/model"
)

func TestExpandDepartmentIDsIncludesSelectedDepartmentsAndDescendants(t *testing.T) {
	departments := []model.Department{
		{ID: 1, ParentID: 0},
		{ID: 2, ParentID: 1},
		{ID: 3, ParentID: 2},
		{ID: 4, ParentID: 0},
	}

	got := expandDepartmentIDs([]uint{2}, departments)
	if want := []uint{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expandDepartmentIDs() = %v, want %v", got, want)
	}
}

func TestExpandDepartmentIDsIgnoresMissingAndDuplicateSelections(t *testing.T) {
	departments := []model.Department{{ID: 2, ParentID: 1}, {ID: 3, ParentID: 2}}

	got := expandDepartmentIDs([]uint{2, 2, 99}, departments)
	if want := []uint{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("expandDepartmentIDs() = %v, want %v", got, want)
	}
}

func TestDeliveryKeyIsStableAndRecipientSpecific(t *testing.T) {
	first := deliveryKey("scheduled_task_run", "run-1", 7)
	second := deliveryKey("scheduled_task_run", "run-1", 7)
	otherRecipient := deliveryKey("scheduled_task_run", "run-1", 8)

	if first != second || first == otherRecipient || len(first) != 64 {
		t.Fatalf("delivery keys first=%q second=%q other=%q", first, second, otherRecipient)
	}
}

func TestBuildDepartmentOptionsKeepsHierarchyAndPromotesOrphans(t *testing.T) {
	departments := []model.Department{
		{ID: 2, Name: "研发", ParentID: 1},
		{ID: 3, Name: "平台组", ParentID: 2},
		{ID: 4, Name: "行政", ParentID: 0},
	}

	got := buildDepartmentOptions(departments)
	if len(got) != 2 || got[0].ID != 2 || len(got[0].Children) != 1 || got[0].Children[0].ID != 3 || got[1].ID != 4 {
		t.Fatalf("department options = %#v", got)
	}
}
