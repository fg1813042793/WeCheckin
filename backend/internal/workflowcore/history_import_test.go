package workflowcore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func historyImportDefinition(t *testing.T, config string) Definition {
	t.Helper()
	definition := validLinearDefinition()
	definition.Form = []FormField{
		{Key: "source", Label: "来源", Type: FormFieldTypeText},
		{Key: "target", Label: "目标", Type: FormFieldTypeTextarea},
		{Key: "amount", Label: "金额", Type: FormFieldTypeAmount},
		{Key: "calc", Label: "计算", Type: FormFieldTypeCalculation, Calculation: &FormCalculation{Expression: "1"}},
		{Key: "label", Label: "标签", Type: FormFieldTypeLabel},
		{Key: "group", Label: "分组", Type: FormFieldTypeGroup, Fields: []FormField{{Key: "grouped", Label: "组内字段", Type: FormFieldTypeText}}},
		{Key: "rows", Label: "来源明细", Type: FormFieldTypeDetailList, Columns: []FormField{
			{Key: "old", Label: "旧列", Type: FormFieldTypeText}, {Key: "sum", Label: "金额", Type: FormFieldTypeNumber},
		}},
		{Key: "newRows", Label: "目标明细", Type: FormFieldTypeDetailList, Columns: []FormField{
			{Key: "new", Label: "新列", Type: FormFieldTypeTextarea}, {Key: "sum", Label: "金额", Type: FormFieldTypeAmount},
		}},
	}
	raw, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["historyImport"] = json.RawMessage(config)
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &definition); err != nil {
		t.Fatal(err)
	}
	return definition
}

func historyImportJSON(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	config, ok := document["historyImport"].(map[string]any)
	if !ok {
		t.Fatalf("historyImport missing from JSON: %s", raw)
	}
	return config
}

func TestHistoryImportRuntimeRoundTripAndDefaults(t *testing.T) {
	definition := historyImportDefinition(t, `{"enabled":true}`)
	raw, err := EncodeRuntimeStartConfig(definition)
	if err != nil {
		t.Fatal(err)
	}
	target := validLinearDefinition()
	if err := ApplyRuntimeStartConfigJSON(&target, raw); err != nil {
		t.Fatal(err)
	}
	config := historyImportJSON(t, target)
	want := map[string]any{"enabled": true, "sourceMode": "select_history", "sourceStatus": "completed", "mappingMode": "same_key", "overwrite": false, "mappings": []any{}}
	if !reflect.DeepEqual(config, want) {
		t.Fatalf("config = %#v, want %#v", config, want)
	}
	if errors := ValidateRuntimeStartConfig(definition); len(errors) != 0 {
		t.Fatalf("default config rejected: %#v", errors)
	}
}

func TestHistoryImportLegacyJSONClearsPreset(t *testing.T) {
	definition := historyImportDefinition(t, `{"enabled":true}`)
	if err := ApplyRuntimeStartConfigJSON(&definition, `{"initiator":{"scope":"all"}}`); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(definition)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"historyImport"`) {
		t.Fatalf("legacy config retained preset: %s", raw)
	}
}

func TestHistoryImportSameKeyIgnoresUnusedMappingDraft(t *testing.T) {
	for _, test := range []struct {
		name, mappings string
	}{
		{"stale fields", `[{"sourceKey":"deleted","targetKey":"removed"}]`},
		{"incompatible fields", `[{"sourceKey":"source","targetKey":"amount"}]`},
		{"stale columns", `[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"deleted","targetKey":"removed"}]}]`},
		{"incompatible columns", `[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"sum"}]}]`},
		{"incomplete mapping", `[{"sourceKey":"","targetKey":""}]`},
		{"duplicate mapping", `[{"sourceKey":"source","targetKey":"target"},{"sourceKey":"source","targetKey":"target"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, mode := range []string{"", HistoryImportMappingSameKey, HistoryImportMappingExplicit} {
				definition := historyImportDefinition(t, fmt.Sprintf(`{"enabled":true,"mappingMode":%q,"mappings":%s}`, mode, test.mappings))
				before, err := json.Marshal(definition.HistoryImport.Mappings)
				if err != nil {
					t.Fatal(err)
				}
				for name, errors := range map[string][]ValidationError{"definition": ValidateDefinition(definition), "runtime": ValidateRuntimeStartConfig(definition)} {
					if invalid := hasValidationCode(errors, ValidationHistoryImport); invalid != (mode == HistoryImportMappingExplicit) {
						t.Errorf("%s mode=%q validation=%#v", name, mode, errors)
					}
				}
				after, err := json.Marshal(definition.HistoryImport.Mappings)
				if err != nil {
					t.Fatal(err)
				}
				if string(before) != string(after) {
					t.Fatal("validation changed the saved mapping draft")
				}
			}
		})
	}
}

func TestHistoryImportSameKeyStillChecksEnumsAndBounds(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*HistoryImportConfig)
	}{
		{"source mode", func(config *HistoryImportConfig) { config.SourceMode = "invalid" }},
		{"source status", func(config *HistoryImportConfig) { config.SourceStatus = "running" }},
		{"mapping count", func(config *HistoryImportConfig) { config.Mappings = make([]HistoryImportMapping, 101) }},
		{"column count", func(config *HistoryImportConfig) {
			config.Mappings = []HistoryImportMapping{{Columns: make([]HistoryImportColumnMapping, 101)}}
		}},
		{"field key length", func(config *HistoryImportConfig) {
			config.Mappings = []HistoryImportMapping{{SourceKey: strings.Repeat("s", 101)}}
		}},
		{"column key length", func(config *HistoryImportConfig) {
			config.Mappings = []HistoryImportMapping{{Columns: []HistoryImportColumnMapping{{TargetKey: strings.Repeat("t", 101)}}}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			definition := historyImportDefinition(t, `{"enabled":true,"mappingMode":"same_key"}`)
			test.change(definition.HistoryImport)
			if !hasValidationCode(ValidateRuntimeStartConfig(definition), ValidationHistoryImport) {
				t.Fatal("same-key mode skipped an active enum or payload bound")
			}
		})
	}
}

func TestHistoryImportValidation(t *testing.T) {
	mapping := func(value string) string { return `{"enabled":true,"mappingMode":"mapping","mappings":` + value + `}` }
	tests := []struct {
		name, config string
		valid        bool
	}{
		{"nil", `null`, true},
		{"same key defaults", `{"enabled":true}`, true},
		{"latest all overwrite", `{"enabled":true,"sourceMode":"latest_history","sourceStatus":"all","overwrite":true}`, true},
		{"text family", mapping(`[{"sourceKey":"source","targetKey":"target"}]`), true},
		{"group business field", mapping(`[{"sourceKey":"grouped","targetKey":"target"}]`), true},
		{"custom columns", mapping(`[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"new"},{"sourceKey":"sum","targetKey":"sum"}]}]`), true},
		{"same detail columns", mapping(`[{"sourceKey":"rows","targetKey":"rows"}]`), true},
		{"source mode", `{"enabled":true,"sourceMode":"newest"}`, false},
		{"status", `{"enabled":true,"sourceStatus":"running"}`, false},
		{"mapping mode", `{"enabled":true,"mappingMode":"auto"}`, false},
		{"empty mappings", mapping(`[]`), false},
		{"missing source", mapping(`[{"sourceKey":"missing","targetKey":"target"}]`), false},
		{"missing target", mapping(`[{"sourceKey":"source","targetKey":"missing"}]`), false},
		{"empty key", mapping(`[{"sourceKey":"","targetKey":"target"}]`), false},
		{"whitespace key", mapping(`[{"sourceKey":" source","targetKey":"target"}]`), false},
		{"long key", mapping(`[{"sourceKey":"` + strings.Repeat("s", 101) + `","targetKey":"target"}]`), false},
		{"duplicate source", mapping(`[{"sourceKey":"source","targetKey":"target"},{"sourceKey":"source","targetKey":"grouped"}]`), false},
		{"duplicate target", mapping(`[{"sourceKey":"source","targetKey":"target"},{"sourceKey":"grouped","targetKey":"target"}]`), false},
		{"incompatible", mapping(`[{"sourceKey":"source","targetKey":"amount"}]`), false},
		{"calculated source", mapping(`[{"sourceKey":"calc","targetKey":"amount"}]`), false},
		{"calculated target", mapping(`[{"sourceKey":"amount","targetKey":"calc"}]`), false},
		{"display field", mapping(`[{"sourceKey":"label","targetKey":"target"}]`), false},
		{"group container", mapping(`[{"sourceKey":"group","targetKey":"target"}]`), false},
		{"scalar columns", mapping(`[{"sourceKey":"source","targetKey":"target","columns":[{"sourceKey":"old","targetKey":"new"}]}]`), false},
		{"explicit empty columns", mapping(`[{"sourceKey":"rows","targetKey":"rows","columns":[]}]`), false},
		{"incompatible default columns", mapping(`[{"sourceKey":"rows","targetKey":"newRows"}]`), false},
		{"missing column", mapping(`[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"missing","targetKey":"new"}]}]`), false},
		{"incompatible columns", mapping(`[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"sum"}]}]`), false},
		{"duplicate source column", mapping(`[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"new"},{"sourceKey":"old","targetKey":"sum"}]}]`), false},
		{"duplicate target column", mapping(`[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"new"},{"sourceKey":"sum","targetKey":"new"}]}]`), false},
		{"disabled stale draft", `{"enabled":false,"mappingMode":"mapping","mappings":[{"sourceKey":"deleted","targetKey":"removed","columns":[{"sourceKey":"gone","targetKey":"gone"}]}]}`, true},
		{"disabled oversized key", `{"enabled":false,"mappings":[{"sourceKey":"` + strings.Repeat("s", 101) + `","targetKey":"target"}]}`, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := historyImportDefinition(t, test.config)
			for name, errors := range map[string][]ValidationError{"definition": ValidateDefinition(definition), "runtime": ValidateRuntimeStartConfig(definition)} {
				invalid := hasValidationCode(errors, "history_import_invalid")
				if invalid == test.valid {
					t.Fatalf("%s validation = %#v, valid=%v", name, errors, test.valid)
				}
			}
		})
	}
}

func TestHistoryImportMappingBounds(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		for _, count := range []int{100, 101} {
			t.Run(fmt.Sprintf("enabled=%v/count=%d", enabled, count), func(t *testing.T) {
				pairs := make([]string, count)
				for i := range pairs {
					pairs[i] = fmt.Sprintf(`{"sourceKey":"f%d","targetKey":"f%d"}`, i, i)
				}
				for _, columns := range []bool{false, true} {
					config := fmt.Sprintf(`{"enabled":%v,"mappingMode":"mapping","mappings":[%s]}`, enabled, strings.Join(pairs, ","))
					if columns {
						config = fmt.Sprintf(`{"enabled":%v,"mappingMode":"mapping","mappings":[{"sourceKey":"rows","targetKey":"rows","columns":[%s]}]}`, enabled, strings.Join(pairs, ","))
					}
					definition := historyImportDefinition(t, config)
					fields := make([]FormField, count)
					for i := range fields {
						fields[i] = FormField{Key: fmt.Sprintf("f%d", i), Label: "字段", Type: FormFieldTypeText}
					}
					if columns {
						definition.Form = []FormField{{Key: "rows", Label: "明细", Type: FormFieldTypeDetailList, Columns: fields}}
					} else {
						definition.Form = fields
					}
					invalid := hasValidationCode(ValidateRuntimeStartConfig(definition), "history_import_invalid")
					if invalid != (count > 100) {
						t.Fatalf("columns=%v count=%d invalid=%v", columns, count, invalid)
					}
				}
			})
		}
	}
}

func TestHistoryImportCloneIsolation(t *testing.T) {
	definition := historyImportDefinition(t, `{"enabled":false,"mappingMode":"mapping","mappings":[{"sourceKey":"rows","targetKey":"newRows","columns":[{"sourceKey":"old","targetKey":"new"}]}]}`)
	before := historyImportJSON(t, definition)
	config := RuntimeStartConfigFromDefinition(definition)
	cloned := CloneHistoryImportConfig(definition.HistoryImport)
	cloned.Mappings[0].Columns[0].TargetKey = "clone_mutation"
	cloned.Mappings[0].SourceKey = "clone_mutation"
	cloned.Enabled = true
	target := validLinearDefinition()
	ApplyRuntimeStartConfig(&target, config)
	config.HistoryImport.Mappings[0].SourceKey = "extraction_mutation"
	config.HistoryImport.Mappings[0].Columns[0].SourceKey = "extraction_mutation"
	config.HistoryImport.Enabled = true
	if target.HistoryImport.Enabled || target.HistoryImport.Mappings[0].SourceKey != "rows" || target.HistoryImport.Mappings[0].Columns[0].SourceKey != "old" {
		t.Fatalf("runtime apply aliases config: %#v", target.HistoryImport)
	}
	target.HistoryImport.Mappings[0].Columns[0].TargetKey = "apply_mutation"
	if !reflect.DeepEqual(before, historyImportJSON(t, definition)) {
		t.Fatal("cloning or applying preset mutated source definition")
	}
	if CloneHistoryImportConfig(nil) != nil {
		t.Fatal("nil preset must stay nil")
	}
}

func TestHistoryImportTypeFamilies(t *testing.T) {
	families := [][]string{
		{FormFieldTypeText, FormFieldTypeTextarea, FormFieldTypePhone, FormFieldTypeEmail},
		{FormFieldTypeNumber, FormFieldTypeAmount}, {FormFieldTypeSelect, FormFieldTypeRadio}, {FormFieldTypeMultiSelect, FormFieldTypeCheckbox},
		{FormFieldTypeDate}, {FormFieldTypeDateTime}, {FormFieldTypeTime}, {FormFieldTypeDateRange},
		{FormFieldTypeUser}, {FormFieldTypeUserMulti}, {FormFieldTypeDepartment}, {FormFieldTypeDepartmentMulti}, {FormFieldTypeAttachment}, {FormFieldTypeBoolean},
	}
	for sourceIndex, sourceFamily := range families {
		for targetIndex, targetFamily := range families {
			for _, sourceType := range sourceFamily {
				for _, targetType := range targetFamily {
					definition := historyImportDefinition(t, `{"enabled":true,"mappingMode":"mapping","mappings":[{"sourceKey":"source","targetKey":"target"}]}`)
					definition.Form[0].Type, definition.Form[1].Type = sourceType, targetType
					invalid := hasValidationCode(ValidateRuntimeStartConfig(definition), ValidationHistoryImport)
					if invalid != (sourceIndex != targetIndex) {
						t.Fatalf("%s -> %s: invalid=%v", sourceType, targetType, invalid)
					}
				}
			}
		}
	}
}

func TestHistoryImportRejectsNestedCalculatedAndUnknownColumns(t *testing.T) {
	for _, fieldType := range []string{FormFieldTypeDetailList, FormFieldTypeCalculation, FormFieldTypeGroup, FormFieldTypeLabel, FormFieldTypeDescription, FormFieldTypeButton, "unknown"} {
		for _, columns := range []string{"", `,"columns":[{"sourceKey":"old","targetKey":"old"}]`} {
			definition := historyImportDefinition(t, `{"enabled":true,"mappingMode":"mapping","mappings":[{"sourceKey":"rows","targetKey":"rows"`+columns+`}]}`)
			definition.Form[6].Columns[0].Type = fieldType
			if !hasValidationCode(ValidateRuntimeStartConfig(definition), ValidationHistoryImport) {
				t.Fatalf("accepted %s detail column (custom=%v)", fieldType, columns != "")
			}
		}
	}
	definition := historyImportDefinition(t, `{"enabled":true,"mappingMode":"mapping","mappings":[{"sourceKey":"source","targetKey":"target"}]}`)
	definition.Form[0].Calculation = &FormCalculation{Expression: "1"}
	if !hasValidationCode(ValidateRuntimeStartConfig(definition), ValidationHistoryImport) {
		t.Fatal("computed field accepted under non-calculation type")
	}
}
