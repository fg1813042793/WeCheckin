package application

import (
	"reflect"
	"testing"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func TestSupplementCommandIncludesVersion(t *testing.T) {
	requireField(t, reflect.TypeOf(SupplementCommand{}), "Version", reflect.TypeOf(uint64(0)), "")
}

func TestFeedbackSummaryIncludesSummaryAndImageCount(t *testing.T) {
	typ := reflect.TypeOf(FeedbackSummary{})
	requireField(t, typ, "Summary", reflect.TypeOf(""), "summary")
	requireField(t, typ, "ImageCount", reflect.TypeOf(int64(0)), "imageCount")
}

func TestUserListQueryFields(t *testing.T) {
	requireExactFields(t, reflect.TypeOf(UserListQuery{}), map[string]reflect.Type{
		"Keyword":  reflect.TypeOf(""),
		"Status":   reflect.TypeOf(domain.Status("")),
		"Page":     reflect.TypeOf(int(0)),
		"PageSize": reflect.TypeOf(int(0)),
	})
}

func TestAdminListQueryFields(t *testing.T) {
	requireExactFields(t, reflect.TypeOf(AdminListQuery{}), map[string]reflect.Type{
		"Keyword":       reflect.TypeOf(""),
		"SubmitterID":   reflect.TypeOf(uint(0)),
		"HandlerID":     reflect.TypeOf((*uint)(nil)),
		"Status":        reflect.TypeOf(domain.Status("")),
		"SubmittedFrom": reflect.TypeOf(int64(0)),
		"SubmittedTo":   reflect.TypeOf(int64(0)),
		"Page":          reflect.TypeOf(int(0)),
		"PageSize":      reflect.TypeOf(int(0)),
	})
}

func requireExactFields(t *testing.T, typ reflect.Type, expected map[string]reflect.Type) {
	t.Helper()
	if typ.NumField() != len(expected) {
		t.Errorf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(expected))
	}
	for name, fieldType := range expected {
		requireField(t, typ, name, fieldType, "")
	}
}

func requireField(t *testing.T, typ reflect.Type, name string, fieldType reflect.Type, jsonName string) {
	t.Helper()
	field, ok := typ.FieldByName(name)
	if !ok {
		t.Errorf("%s.%s is missing", typ.Name(), name)
		return
	}
	if field.Type != fieldType {
		t.Errorf("%s.%s type = %s, want %s", typ.Name(), name, field.Type, fieldType)
	}
	if jsonName != "" && field.Tag.Get("json") != jsonName {
		t.Errorf("%s.%s json tag = %q, want %q", typ.Name(), name, field.Tag.Get("json"), jsonName)
	}
}
