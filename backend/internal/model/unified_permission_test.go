package model

import (
	"reflect"
	"strings"
	"testing"
)

func TestUnifiedPermissionModelsUseExpectedTables(t *testing.T) {
	if (Permission{}).TableName() != "permissions" {
		t.Fatalf("Permission table name = %s", (Permission{}).TableName())
	}
	if (PermissionGrant{}).TableName() != "permission_grants" {
		t.Fatalf("PermissionGrant table name = %s", (PermissionGrant{}).TableName())
	}
}

func TestPermissionModelHasStableColumns(t *testing.T) {
	typ := reflect.TypeOf(Permission{})
	checks := map[string][]string{
		"Key":        {`json:"key"`, "column:permission_key", "uniqueIndex"},
		"Platform":   {`json:"platform"`, "column:permission_platform", "index"},
		"Type":       {`json:"type"`, "column:permission_type", "index"},
		"ParentKey":  {`json:"parentKey"`, "column:permission_parent_key"},
		"ResourceID": {`json:"resourceId"`, "column:permission_resource_id"},
		"Icon":       {`json:"icon"`, "column:permission_icon"},
		"Perms":      {`json:"perms"`, "column:permission_perms"},
	}
	for name, snippets := range checks {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("Permission must include field %s", name)
		}
		tag := string(field.Tag)
		for _, snippet := range snippets {
			if !strings.Contains(tag, snippet) {
				t.Fatalf("Permission.%s tag must include %s, got %s", name, snippet, tag)
			}
		}
	}
}

func TestPermissionGrantModelSupportsRoleAndUserSubjects(t *testing.T) {
	typ := reflect.TypeOf(PermissionGrant{})
	checks := map[string][]string{
		"SubjectType":   {`json:"subjectType"`, "column:grant_subject_type", "idx_permission_grants_subject_permission"},
		"SubjectID":     {`json:"subjectId"`, "column:grant_subject_id", "idx_permission_grants_subject_permission"},
		"PermissionKey": {`json:"permissionKey"`, "column:grant_permission_key", "idx_permission_grants_subject_permission"},
		"PermissionID":  {`json:"permissionId"`, "column:grant_permission_id"},
		"Effect":        {`json:"effect"`, "column:grant_effect"},
		"ScopeValue":    {`json:"scopeValue"`, "column:grant_scope_value"},
	}
	for name, snippets := range checks {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("PermissionGrant must include field %s", name)
		}
		tag := string(field.Tag)
		for _, snippet := range snippets {
			if !strings.Contains(tag, snippet) {
				t.Fatalf("PermissionGrant.%s tag must include %s, got %s", name, snippet, tag)
			}
		}
	}
}
