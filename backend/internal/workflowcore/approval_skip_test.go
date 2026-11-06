package workflowcore

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestApprovalSkipJSONAndBPMN(t *testing.T) {
	for _, setting := range []string{`{}`, `{"skipIfNoAssignee":false}`, `{"skipIfNoAssignee":true}`} {
		t.Run(setting, func(t *testing.T) {
			definition := validLinearDefinition()
			if err := json.Unmarshal([]byte(setting), &definition.Nodes[1]); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(definition)
			if err != nil {
				t.Fatal(err)
			}
			wantEnabled := strings.Contains(setting, "true")
			if got := strings.Contains(string(encoded), `"skipIfNoAssignee":true`); got != wantEnabled {
				t.Fatalf("skip setting lost in JSON: %s", encoded)
			}
			if !wantEnabled && strings.Contains(string(encoded), "skipIfNoAssignee") {
				t.Fatal("false/default setting must be omitted")
			}
			var decoded Definition
			if err := json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			bpmn, err := CompileBPMN(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Contains(string(bpmn), `flowable:skipIfNoAssignee="true"`); got != wantEnabled {
				t.Fatalf("BPMN skip attribute = %v, want %v: %s", got, wantEnabled, bpmn)
			}
			if !wantEnabled && strings.Contains(string(bpmn), "skipIfNoAssignee") {
				t.Fatal("default BPMN must omit skip attribute")
			}
		})
	}
}

func TestApprovalSkipValidation(t *testing.T) {
	for _, nodeType := range []string{NodeTypeStart, NodeTypeEnd, NodeTypeApproval, NodeTypeHandle, NodeTypeCC, NodeTypeNotify, NodeTypeAutomation, NodeTypeTimer, NodeTypeExclusive, NodeTypeParallel} {
		t.Run(nodeType, func(t *testing.T) {
			definition := validLinearDefinition()
			node := &definition.Nodes[1]
			node.Type = nodeType
			if err := json.Unmarshal([]byte(`{"skipIfNoAssignee":true}`), node); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, issue := range ValidateDefinition(definition) {
				if issue.Code == "skip_if_no_assignee_invalid" && issue.NodeID == node.ID {
					found = true
				}
			}
			if found != (nodeType != NodeTypeApproval) {
				t.Fatalf("wrong-node validation = %v for %s", found, nodeType)
			}
		})
	}
	for _, assignee := range []*Assignee{nil, {}, {Type: AssigneeTypeUser, Value: " \t"}, {Type: "unknown", Value: "user"}} {
		definition := validLinearDefinition()
		if err := json.Unmarshal([]byte(`{"skipIfNoAssignee":true}`), &definition.Nodes[1]); err != nil {
			t.Fatal(err)
		}
		definition.Nodes[1].Assignee = assignee
		if issues := ValidateDefinition(definition); !hasValidationCode(issues, ValidationAssigneeRequired) {
			t.Fatalf("skip must still require legal assignee source: %#v", assignee)
		}
	}
}
