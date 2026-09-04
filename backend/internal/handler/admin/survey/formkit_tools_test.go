package survey

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormKitToolResponsesKeepStableJSONFields(t *testing.T) {
	tests := []struct {
		name string
		body any
		want []string
	}{
		{name: "eval", body: EvalExprResponse{}, want: []string{`"value":null`}},
		{name: "validate", body: ValidateAnswersResponse{}, want: []string{`"ok":false`, `"errors":null`}},
		{name: "apply", body: ApplyFormResponse{}, want: []string{`"answers":null`, `"states":null`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatal(err)
			}
			text := string(encoded)
			for _, field := range tt.want {
				if !strings.Contains(text, field) {
					t.Fatalf("JSON %s missing %s", text, field)
				}
			}
		})
	}
}

func TestFieldValidationErrorKeepsStableJSONFields(t *testing.T) {
	encoded, err := json.Marshal(FieldValidationError{QuestionID: "q1", Message: "required"})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"questionId":"q1","message":"required"}` {
		t.Fatalf("unexpected JSON: %s", encoded)
	}
}
