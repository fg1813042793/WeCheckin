package application

import (
	"encoding/json"
	"reflect"
	"testing"

	"wecheckin/backend/internal/modules/userfeedback/domain"
)

func TestSupplementCommandIncludesVersion(t *testing.T) {
	requireField(t, reflect.TypeOf(SupplementCommand{}), "Version", reflect.TypeOf(uint64(0)), "")
}

func TestFeedbackSummaryIncludesListPresentationFields(t *testing.T) {
	typ := reflect.TypeOf(FeedbackSummary{})
	requireField(t, typ, "Summary", reflect.TypeOf(""), "summary")
	requireField(t, typ, "ImageCount", reflect.TypeOf(int64(0)), "imageCount")
	requireField(t, typ, "FirstImageURL", reflect.TypeOf(""), "firstImageUrl,omitempty")
}

func TestFeedbackResponseIDsMarshalAsDecimalStrings(t *testing.T) {
	payload := FeedbackDetail{
		FeedbackSummary: FeedbackSummary{ID: 9007199254740993},
		Messages: []Message{{
			ID:          9007199254740995,
			Attachments: []Attachment{{ID: 9007199254740997}},
		}},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal feedback detail: %v", err)
	}
	var decoded struct {
		ID       string `json:"id"`
		Messages []struct {
			ID          string `json:"id"`
			Attachments []struct {
				ID string `json:"id"`
			} `json:"attachments"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal feedback detail: %v", err)
	}
	if decoded.ID != "9007199254740993" || decoded.Messages[0].ID != "9007199254740995" || decoded.Messages[0].Attachments[0].ID != "9007199254740997" {
		t.Fatalf("feedback response IDs lost decimal precision: %#v", decoded)
	}

	for _, contract := range []struct {
		typ  reflect.Type
		name string
	}{
		{typ: reflect.TypeOf(FeedbackSummary{}), name: "FeedbackSummary"},
		{typ: reflect.TypeOf(Message{}), name: "Message"},
		{typ: reflect.TypeOf(Attachment{}), name: "Attachment"},
	} {
		field, ok := contract.typ.FieldByName("ID")
		if !ok {
			t.Fatalf("%s.ID is missing", contract.name)
		}
		if field.Tag.Get("json") != "id,string" {
			t.Errorf("%s.ID json tag = %q, want %q", contract.name, field.Tag.Get("json"), "id,string")
		}
		if field.Tag.Get("swaggertype") != "string" {
			t.Errorf("%s.ID swaggertype = %q, want string", contract.name, field.Tag.Get("swaggertype"))
		}
	}
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
