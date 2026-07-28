package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestPositionModelAndUserPositionField(t *testing.T) {
	positionType := reflect.TypeOf(Position{})
	requiredFields := map[string]string{
		"Name":   "position_name",
		"Sort":   "position_sort",
		"Status": "position_status",
	}
	for fieldName, column := range requiredFields {
		field, ok := positionType.FieldByName(fieldName)
		if !ok {
			t.Fatalf("Position must include field %s", fieldName)
		}
		if !strings.Contains(string(field.Tag), "column:"+column) {
			t.Fatalf("Position.%s must map to column %s", fieldName, column)
		}
	}

	userField, ok := reflect.TypeOf(User{}).FieldByName("PositionID")
	if !ok {
		t.Fatalf("User must include PositionID")
	}
	tag := string(userField.Tag)
	for _, snippet := range []string{`json:"positionId"`, "column:user_position_id", "index"} {
		if !strings.Contains(tag, snippet) {
			t.Fatalf("User.PositionID tag must include %s, got %s", snippet, tag)
		}
	}
}
