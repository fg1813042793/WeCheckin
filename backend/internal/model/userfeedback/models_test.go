package userfeedbackmodel

import (
	"reflect"
	"strings"
	"testing"
)

func TestFeedbackStatusAndVersionDefaultsMatchMigration(t *testing.T) {
	typ := reflect.TypeOf(Feedback{})
	requireGORMTagParts(t, typ, "Status", "column:feedback_status", "size:24", "default:pending")
	requireGORMTagParts(t, typ, "Version", "column:version", "default:1")
}

func requireGORMTagParts(t *testing.T, typ reflect.Type, fieldName string, expected ...string) {
	t.Helper()
	field, ok := typ.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s.%s is missing", typ.Name(), fieldName)
	}
	parts := make(map[string]bool)
	for _, part := range strings.Split(field.Tag.Get("gorm"), ";") {
		parts[part] = true
	}
	for _, part := range expected {
		if !parts[part] {
			t.Errorf("%s.%s gorm tag %q is missing %q", typ.Name(), fieldName, field.Tag.Get("gorm"), part)
		}
	}
}
