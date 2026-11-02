package infrastructure

import (
	"encoding/json"
	"strings"
	"testing"

	workflowmodel "wecheckin/backend/internal/model/workflow"
	"wecheckin/backend/internal/workflowcore"
)

func TestPublishedDefinitionIncludesHistoryImportInListAndDetail(t *testing.T) {
	var definition workflowcore.Definition
	if err := json.Unmarshal([]byte(`{"historyImport":{"enabled":true,"sourceMode":"latest_history","sourceStatus":"completed","mappingMode":"mapping","overwrite":true,"mappings":[{"sourceKey":"rows","targetKey":"rows","columns":[{"sourceKey":"old","targetKey":"new"}]}]}}`), &definition); err != nil {
		t.Fatal(err)
	}
	for _, includeGraph := range []bool{false, true} {
		result := publishedDefinition(workflowmodel.Definition{}, definition, 2, publishedAssigneeLabels{}, includeGraph)
		raw, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		for _, snippet := range []string{`"historyImport"`, `"sourceMode":"latest_history"`, `"overwrite":true`, `"columns":[{"sourceKey":"old","targetKey":"new"}]`} {
			if !strings.Contains(string(raw), snippet) {
				t.Fatalf("published graph=%v missing %s: %s", includeGraph, snippet, raw)
			}
		}
		result.HistoryImport.Mappings[0].SourceKey = "mutated"
		result.HistoryImport.Mappings[0].Columns[0].SourceKey = "mutated"
		if definition.HistoryImport.Mappings[0].SourceKey != "rows" || definition.HistoryImport.Mappings[0].Columns[0].SourceKey != "old" {
			t.Fatal("published DTO aliases source definition")
		}
	}
}

func TestPublishedDefinitionHistoryImportDefaultsToCompleted(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		definition := workflowcore.Definition{HistoryImport: &workflowcore.HistoryImportConfig{Enabled: enabled}}
		for _, includeGraph := range []bool{false, true} {
			result := publishedDefinition(workflowmodel.Definition{}, definition, 2, publishedAssigneeLabels{}, includeGraph)
			if result.HistoryImport == nil || result.HistoryImport.SourceStatus != workflowcore.HistoryImportStatusCompleted {
				t.Fatalf("enabled=%v includeGraph=%v preset=%#v", enabled, includeGraph, result.HistoryImport)
			}
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(raw), `"sourceStatus":"completed"`) {
				t.Fatalf("published JSON omitted completed default: %s", raw)
			}
		}
		if definition.HistoryImport.SourceStatus != "" {
			t.Fatal("DTO normalization mutated stored definition")
		}
	}
}
