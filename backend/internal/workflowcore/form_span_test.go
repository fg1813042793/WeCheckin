package workflowcore

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestDefinitionFormSpanJSONRoundTrip(t *testing.T) {
	spans := []string{"0", "4.8"}
	for value := 1; value <= 24; value++ {
		spans = append(spans, fmt.Sprint(value))
	}
	for _, span := range spans {
		for _, context := range []string{"root", "group", "detail", "group_detail"} {
			t.Run(context+"/"+span, func(t *testing.T) {
				definition := validLinearDefinition()
				definition.Form = formSpanFixture(t, span, context)
				if errors := ValidateDefinition(definition); len(errors) > 0 {
					t.Fatalf("span %s in %s rejected: %#v", span, context, errors)
				}
				payload, err := json.Marshal(definition)
				if err != nil {
					t.Fatal(err)
				}
				var reloaded Definition
				if err := json.Unmarshal(payload, &reloaded); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(definition, reloaded) {
					t.Fatalf("span %s changed after JSON reload", span)
				}
			})
		}
	}
}

func TestDefinitionRejectsUnsupportedFormSpans(t *testing.T) {
	for _, span := range []string{"-1", "4.7", "4.9", "4.79", "4.800001", "24.1", "25", "120"} {
		for _, context := range []string{"root", "group", "detail", "group_detail"} {
			t.Run(context+"/"+span, func(t *testing.T) {
				definition := validLinearDefinition()
				definition.Form = formSpanFixture(t, span, context)
				if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationFormFieldSpan) {
					t.Fatalf("span %s in %s should be rejected, got %#v", span, context, errors)
				}
			})
		}
	}
}

func TestDefinitionOmittedFormSpanRemainsDefault(t *testing.T) {
	var field FormField
	if err := json.Unmarshal([]byte(`{"key":"legacy","label":"旧字段","type":"text"}`), &field); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(field)
	if err != nil {
		t.Fatal(err)
	}
	var encoded map[string]json.RawMessage
	if err := json.Unmarshal(payload, &encoded); err != nil {
		t.Fatal(err)
	}
	if _, exists := encoded["span"]; exists || field.Span != 0 {
		t.Fatalf("legacy default width must remain omitted: %s", payload)
	}
}

func formSpanFixture(t *testing.T, span, context string) []FormField {
	t.Helper()
	field := fmt.Sprintf(`{"key":"value","label":"字段","type":"text","span":%s}`, span)
	if context == "detail" || context == "group_detail" {
		field = `{"key":"items","label":"明细","type":"detail_list","columns":[` + field + `]}`
	}
	if context == "group" || context == "group_detail" {
		field = `{"key":"section","label":"分组","type":"group","fields":[` + field + `]}`
	}
	var fields []FormField
	if err := json.Unmarshal([]byte("["+field+"]"), &fields); err != nil {
		t.Fatalf("decode span %s in %s: %v", span, context, err)
	}
	return fields
}
