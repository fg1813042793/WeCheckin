package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestRoleModelHasUnifiedAuditFields(t *testing.T) {
	roleType := reflect.TypeOf(Role{})
	required := map[string]string{
		"CreateBy":     "column:create_by",
		"UpdateBy":     "column:update_by",
		"CreateDeptID": "column:create_dept_id",
		"UpdateDeptID": "column:update_dept_id",
	}
	for field, columnTag := range required {
		structField, ok := roleType.FieldByName(field)
		if !ok {
			t.Fatalf("Role missing unified audit field %s", field)
		}
		if !strings.Contains(string(structField.Tag), columnTag) {
			t.Fatalf("Role.%s must map %s, got tag %q", field, columnTag, structField.Tag)
		}
	}
}
