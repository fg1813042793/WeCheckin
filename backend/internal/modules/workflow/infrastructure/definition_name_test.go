package infrastructure

import (
	"reflect"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/workflowcore"
)

func TestPublishedDefinitionUsesCurrentManagementNameWithoutChangingPublishedForm(t *testing.T) {
	definition := workflowcore.Definition{Name: "旧管理名称", DisplayName: "绩效考评单", Form: []workflowcore.FormField{{Key: "historic", Label: "历史字段", Type: workflowcore.FormFieldTypeText}}}
	for _, graph := range []bool{false, true} {
		got := publishedDefinition(workflowmodel.Definition{ID: 7, Name: " MH-通用绩效考评单 "}, definition, 16, publishedAssigneeLabels{}, graph)
		if got.Name != "MH-通用绩效考评单" || got.DisplayName != "绩效考评单" || got.Version != 16 || !reflect.DeepEqual(got.Form, definition.Form) {
			t.Fatalf("name must be current metadata while form stays version-bound: %+v", got)
		}
	}
	if definition.Name != "旧管理名称" {
		t.Fatal("published definition was modified")
	}
	got := publishedDefinition(workflowmodel.Definition{}, definition, 16, publishedAssigneeLabels{}, false)
	if got.Name != "旧管理名称" {
		t.Fatalf("missing current name fallback = %q", got.Name)
	}
}
