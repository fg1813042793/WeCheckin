package workflowservice

import (
	"encoding/json"
	"strings"
	"testing"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/workflowcore"
)

func TestHistoryImportDraftNormalizationAndRevision(t *testing.T) {
	definition := newDefaultDefinition("history_import", "历史导入")
	original, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(original, &document); err != nil {
		t.Fatal(err)
	}
	document["historyImport"] = json.RawMessage(`{"enabled":true}`)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	normalized, encoded, err := normalizeDraft(raw, definition.Key, definition.Name)
	if err != nil {
		t.Fatalf("history import draft rejected: %v", err)
	}
	for _, snippet := range []string{`"historyImport"`, `"sourceMode":"select_history"`, `"sourceStatus":"completed"`, `"mappingMode":"same_key"`} {
		if !strings.Contains(encoded, snippet) {
			t.Fatalf("normalized draft missing %s: %s", snippet, encoded)
		}
	}
	previousConfig, err := workflowcore.EncodeRuntimeStartConfig(definition)
	if err != nil {
		t.Fatal(err)
	}
	nextConfig, err := workflowcore.EncodeRuntimeStartConfig(normalized)
	if err != nil {
		t.Fatal(err)
	}
	item := model.WorkflowDefinition{Name: definition.Name, DraftJSON: encoded, StartConfigJSON: previousConfig, StartConfigRevision: 4, CurrentVersion: 2}
	updates := definitionContentUpdates(item, item.Name, item.DisplayName, item.Description, item.Category, item.Status, encoded, nextConfig, nil)
	if len(updates) != 2 || updates["definition_start_config_json"] != nextConfig || updates["definition_start_config_revision"] != 5 {
		t.Fatalf("preset must update only live config and revision: %#v", updates)
	}
	item.StartConfigJSON = nextConfig
	item.StartConfigRevision = 5
	if updates := definitionContentUpdates(item, item.Name, item.DisplayName, item.Description, item.Category, item.Status, encoded, nextConfig, nil); len(updates) != 0 {
		t.Fatalf("unchanged save: %#v", updates)
	}
	if err := workflowcore.ApplyRuntimeStartConfigJSON(&definition, nextConfig); err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(versionSnapshot{Definition: definition}, versionSnapshot{Definition: normalized}) {
		t.Fatal("live preset alone must not create a publish difference")
	}
}

func TestHistoryImportPublishedFieldsValidation(t *testing.T) {
	published := newDefaultDefinition("history_import", "历史导入")
	published.Form = []workflowcore.FormField{{Key: "existing", Label: "已发布字段", Type: workflowcore.FormFieldTypeText}}
	before, err := json.Marshal(published)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"sourceKey", "targetKey"} {
		for _, enabled := range []bool{true, false} {
			config := map[string]any{"enabled": enabled, "mappingMode": "mapping", "mappings": []map[string]string{{"sourceKey": "existing", "targetKey": "existing"}}}
			config["mappings"].([]map[string]string)[0][field] = "unpublished"
			raw, err := json.Marshal(map[string]any{"historyImport": config, "initiator": map[string]string{"scope": "all"}})
			if err != nil {
				t.Fatal(err)
			}
			current := published
			if err := workflowcore.ApplyRuntimeStartConfigJSON(&current, string(raw)); err != nil {
				t.Fatal(err)
			}
			errors := workflowcore.ValidateRuntimeStartConfig(current)
			if enabled && (len(errors) == 0 || errors[0].Code != "history_import_invalid") {
				t.Fatalf("enabled %s reference must reject unpublished fields: %#v", field, errors)
			}
			if !enabled && len(errors) != 0 {
				t.Fatalf("disabled stale %s blocked save: %#v", field, errors)
			}
		}
	}
	after, err := json.Marshal(published)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("runtime overlay mutated immutable published snapshot")
	}
}

func TestHistoryImportDraftRejectsWrongJSONTypesAndUnknownFields(t *testing.T) {
	for _, config := range []string{
		`{"enabled":"true"}`, `{"overwrite":1}`, `{"mappings":{}}`,
		`{"mappings":[{"sourceKey":7,"targetKey":"x"}]}`,
		`{"mappings":[{"sourceKey":"x","targetKey":"x","columns":{}}]}`,
		`{"enabled":true,"unknown":1}`,
		`{"mappings":[{"sourceKey":"x","targetKey":"x","columns":[{"sourceKey":"x","targetKey":"x","columns":[]}]}]}`,
	} {
		if _, _, err := normalizeDraft(json.RawMessage(`{"historyImport":`+config+`}`), "history_import", "历史导入"); err == nil {
			t.Fatalf("accepted malformed config: %s", config)
		}
	}
}
