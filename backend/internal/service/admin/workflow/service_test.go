package workflowservice

import (
	"encoding/json"
	"testing"

	workflowcore "wecheckin/backend/internal/workflow"
)

func TestNewDefaultDefinitionCreatesConnectedStartAndEnd(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")
	if definition.Key != "purchase_approval" || definition.Name != "采购申请审批" {
		t.Fatalf("unexpected identity: %#v", definition)
	}
	if len(definition.Nodes) != 2 || definition.Nodes[0].Type != workflowcore.NodeTypeStart || definition.Nodes[1].Type != workflowcore.NodeTypeEnd {
		t.Fatalf("unexpected nodes: %#v", definition.Nodes)
	}
	if len(definition.Edges) != 1 || definition.Edges[0].Source != "start" || definition.Edges[0].Target != "end" {
		t.Fatalf("unexpected edges: %#v", definition.Edges)
	}
}

func TestNormalizeDraftUsesStoredIdentityAndPreservesNumbers(t *testing.T) {
	raw := json.RawMessage(`{
		"schemaVersion":1,
		"key":"client-side-key",
		"name":"客户端名称",
		"nodes":[
			{"id":"start","type":"start","name":"开始","position":{"x":120,"y":48}},
			{"id":"gateway","type":"exclusive","name":"金额判断","gatewayMode":"split"},
			{"id":"end","type":"end","name":"结束"}
		],
		"edges":[
			{"id":"a","source":"start","target":"gateway"},
			{"id":"b","source":"gateway","target":"end","condition":{"field":"amount","operator":"gte","value":10000000000000001}},
			{"id":"c","source":"gateway","target":"end","default":true}
		]
	}`)

	definition, encoded, err := normalizeDraft(raw, "purchase_approval", "采购申请审批")
	if err != nil {
		t.Fatalf("normalize draft: %v", err)
	}
	if definition.Key != "purchase_approval" || definition.Name != "采购申请审批" {
		t.Fatalf("identity was not normalized: %#v", definition)
	}
	if definition.Nodes[0].Position == nil || definition.Nodes[0].Position.X != 120 || definition.Nodes[0].Position.Y != 48 {
		t.Fatalf("node position was not preserved: %#v", definition.Nodes[0])
	}
	value, ok := definition.Edges[1].Condition.Value.(json.Number)
	if !ok || value.String() != "10000000000000001" {
		t.Fatalf("condition number lost precision: %#v", definition.Edges[1].Condition.Value)
	}

	var persisted workflowcore.Definition
	if err := json.Unmarshal([]byte(encoded), &persisted); err != nil {
		t.Fatalf("persisted draft is invalid JSON: %v", err)
	}
	if persisted.Key != "purchase_approval" || persisted.Name != "采购申请审批" {
		t.Fatalf("persisted identity mismatch: %#v", persisted)
	}
	if persisted.Nodes[0].Position == nil || persisted.Nodes[0].Position.X != 120 || persisted.Nodes[0].Position.Y != 48 {
		t.Fatalf("persisted node position mismatch: %#v", persisted.Nodes[0])
	}
}
