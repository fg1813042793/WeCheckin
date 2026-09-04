package workflowservice

import (
	"encoding/json"
	"strings"
	"testing"

	"gorm.io/gorm"

	"wecheckin/backend/internal/model"
	"wecheckin/backend/internal/workflowcore"
)

func TestDefinitionContentUpdatesOnlyIncludesChangedColumns(t *testing.T) {
	item := model.WorkflowDefinition{
		Name:        "请假审批",
		Description: "员工请假",
		Category:    "人事",
		Status:      model.DefinitionStatusDraft,
		DraftJSON:   `{"schemaVersion":1}`,
	}

	updates := definitionContentUpdates(item, "请假审批", "员工请假", "人事", model.DefinitionStatusDraft, `{"schemaVersion":1,"name":"请假审批"}`, nil)
	if len(updates) != 1 || updates["definition_draft_json"] == nil {
		t.Fatalf("only changed draft should be updated, got %#v", updates)
	}
	for _, unchanged := range []string{"definition_name", "definition_description", "definition_category", "definition_status"} {
		if _, exists := updates[unchanged]; exists {
			t.Fatalf("unchanged column %s should not be updated: %#v", unchanged, updates)
		}
	}
}

func TestDefinitionContentUpdatesSkipsUnchangedSave(t *testing.T) {
	item := model.WorkflowDefinition{
		Name:        "请假审批",
		Description: "员工请假",
		Category:    "人事",
		Status:      model.DefinitionStatusDraft,
		DraftJSON:   `{"schemaVersion":1}`,
	}

	updates := definitionContentUpdates(item, item.Name, item.Description, item.Category, item.Status, item.DraftJSON, nil)
	if len(updates) != 0 {
		t.Fatalf("unchanged save should not write database, got %#v", updates)
	}
}

func TestDefinitionContentUpdatesHandlesOptionalLogoChange(t *testing.T) {
	item := model.WorkflowDefinition{LogoURL: "/uploads/workflow-logos/old.png"}

	replacement := "/uploads/workflow-logos/new.png"
	updates := definitionContentUpdates(item, item.Name, item.Description, item.Category, item.Status, item.DraftJSON, &replacement)
	if updates["definition_logo_url"] != replacement {
		t.Fatalf("replacement logo update = %#v", updates)
	}

	removed := ""
	updates = definitionContentUpdates(item, item.Name, item.Description, item.Category, item.Status, item.DraftJSON, &removed)
	if value, exists := updates["definition_logo_url"]; !exists || value != "" {
		t.Fatalf("removed logo update = %#v", updates)
	}
}

func TestDefinitionUpdateSessionSkipsRedundantDefaultTransaction(t *testing.T) {
	db := &gorm.DB{Config: &gorm.Config{}}
	if session := definitionUpdateSession(db); !session.SkipDefaultTransaction {
		t.Fatal("single-statement workflow definition update should skip GORM default transaction")
	}
}

func TestCopyCreateRequestUsesOnlySourceDraftAndNewMetadata(t *testing.T) {
	source := model.WorkflowDefinition{
		Key:            "leave_v1",
		Name:           "请假审批",
		Description:    "原流程说明",
		Category:       "人事",
		LogoURL:        "/uploads/workflow-logos/source.png",
		CurrentVersion: 8,
		DraftJSON:      `{"schemaVersion":1,"key":"leave_v1","name":"请假审批","nodes":[{"id":"start","type":"start","name":"开始"},{"id":"end","type":"end","name":"结束"}],"edges":[{"id":"flow","source":"start","target":"end"}]}`,
	}
	request := CopyRequest{
		Key:         "leave_v2",
		Name:        "新请假审批",
		Description: "新流程说明",
		Category:    "行政",
		LogoURL:     "/uploads/workflow-logos/new.png",
	}

	createRequest := createRequestForCopy(source, request)
	if createRequest.Key != request.Key || createRequest.Name != request.Name ||
		createRequest.Description != request.Description || createRequest.Category != request.Category ||
		createRequest.LogoURL != request.LogoURL {
		t.Fatalf("copy must use new metadata, got %#v", createRequest)
	}
	if string(createRequest.Draft) != source.DraftJSON {
		t.Fatalf("copy must preserve only the source draft, got %s", createRequest.Draft)
	}
}

func TestDefinitionModelForCreateAlwaysStartsAsUnpublishedDraft(t *testing.T) {
	request := CreateRequest{Key: "leave_v2", Name: "新请假审批", Description: "说明", Category: "人事"}
	item := definitionModelForCreate(66, request, `{"schemaVersion":1}`, "/uploads/workflow-logos/new.png", 123456)

	if item.Status != model.DefinitionStatusDraft || item.CurrentVersion != 0 {
		t.Fatalf("copied definition must start as an unpublished draft, got status=%d currentVersion=%d", item.Status, item.CurrentVersion)
	}
	if item.AddUserID != 66 || item.EditUserID != 66 || item.AddTime != 123456 || item.EditTime != 123456 {
		t.Fatalf("copied definition audit fields are invalid: %#v", item)
	}
}

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

func TestNormalizeDraftMigratesLegacyTopLevelInitiatorToStartNode(t *testing.T) {
	raw := json.RawMessage(`{
		"schemaVersion":1,
		"key":"client-side-key",
		"name":"客户端名称",
		"initiator":{"scope":"specified","userIds":[7,8],"departmentIds":[3,5],"excludedUserIds":[9]},
		"nodes":[
			{"id":"start","type":"start","name":"开始"},
			{"id":"end","type":"end","name":"结束"}
		],
		"edges":[{"id":"a","source":"start","target":"end"}]
	}`)

	definition, encoded, err := normalizeDraft(raw, "purchase_approval", "采购申请审批")
	if err != nil {
		t.Fatalf("normalize legacy draft: %v", err)
	}
	if definition.Nodes[0].Initiator == nil || definition.Nodes[0].Initiator.Scope != workflowcore.InitiatorScopeSpecified {
		t.Fatalf("legacy initiator was not migrated to start node: %#v", definition.Nodes[0])
	}
	if got := definition.Nodes[0].Initiator.UserIDs; len(got) != 2 || got[0] != 7 || got[1] != 8 {
		t.Fatalf("legacy initiator users were not preserved: %#v", got)
	}
	if got := definition.Nodes[0].Initiator.DepartmentIDs; len(got) != 2 || got[0] != 3 || got[1] != 5 {
		t.Fatalf("legacy initiator departments were not preserved: %#v", got)
	}
	if got := definition.Nodes[0].Initiator.ExcludedUserIDs; len(got) != 1 || got[0] != 9 {
		t.Fatalf("legacy initiator exclusions were not preserved: %#v", got)
	}
	var persisted map[string]interface{}
	if err := json.Unmarshal([]byte(encoded), &persisted); err != nil {
		t.Fatalf("persisted draft is invalid JSON: %v", err)
	}
	if _, ok := persisted["initiator"]; ok {
		t.Fatalf("legacy top-level initiator should not be persisted: %s", encoded)
	}
}

func TestNormalizeDraftStillRejectsOtherUnknownFields(t *testing.T) {
	raw := json.RawMessage(`{
		"schemaVersion":1,
		"key":"client-side-key",
		"name":"客户端名称",
		"unexpected":true,
		"nodes":[
			{"id":"start","type":"start","name":"开始"},
			{"id":"end","type":"end","name":"结束"}
		],
		"edges":[{"id":"a","source":"start","target":"end"}]
	}`)

	if _, _, err := normalizeDraft(raw, "purchase_approval", "采购申请审批"); err == nil || !strings.Contains(err.Error(), `unknown field "unexpected"`) {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestApplyPublishInitiatorDefaultsMissingConfigurationToAllUsers(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")

	applyPublishInitiator(&definition, nil)

	initiator := definition.Nodes[0].Initiator
	if initiator == nil || initiator.Scope != workflowcore.InitiatorScopeAll || len(initiator.UserIDs) != 0 || len(initiator.DepartmentIDs) != 0 {
		t.Fatalf("missing publish configuration should default to all users, got %#v", initiator)
	}
}

func TestApplyPublishInitiatorPreservesDraftConfigurationWhenRequestIsOmitted(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{3, 5}, DepartmentIDs: []uint{11, 12}, ExcludedUserIDs: []uint{8},
	}

	applyPublishInitiator(&definition, nil)

	initiator := definition.Nodes[0].Initiator
	if initiator == nil || initiator.Scope != workflowcore.InitiatorScopeSpecified ||
		len(initiator.UserIDs) != 2 || initiator.UserIDs[0] != 3 || initiator.UserIDs[1] != 5 ||
		len(initiator.DepartmentIDs) != 2 || initiator.DepartmentIDs[0] != 11 || initiator.DepartmentIDs[1] != 12 ||
		len(initiator.ExcludedUserIDs) != 1 || initiator.ExcludedUserIDs[0] != 8 {
		t.Fatalf("omitted publish configuration should preserve the draft value, got %#v", initiator)
	}
}

func TestApplyPublishInitiatorOverridesStartNodeWithDefensiveCopy(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")
	requested := &workflowcore.InitiatorConfig{
		Scope: workflowcore.InitiatorScopeSpecified, UserIDs: []uint{7, 8}, DepartmentIDs: []uint{3, 5}, ExcludedUserIDs: []uint{8, 9},
	}

	applyPublishInitiator(&definition, requested)
	requested.UserIDs[0] = 99
	requested.DepartmentIDs[0] = 99
	requested.ExcludedUserIDs[0] = 99

	initiator := definition.Nodes[0].Initiator
	if initiator == nil || initiator.Scope != workflowcore.InitiatorScopeSpecified ||
		len(initiator.UserIDs) != 2 || initiator.UserIDs[0] != 7 || initiator.UserIDs[1] != 8 ||
		len(initiator.DepartmentIDs) != 2 || initiator.DepartmentIDs[0] != 3 || initiator.DepartmentIDs[1] != 5 ||
		len(initiator.ExcludedUserIDs) != 2 || initiator.ExcludedUserIDs[0] != 8 || initiator.ExcludedUserIDs[1] != 9 {
		t.Fatalf("publish configuration was not copied into the start node: %#v", initiator)
	}
}

func TestApplyPublishInitiatorLeavesInvalidSpecifiedUsersForDefinitionValidation(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")
	applyPublishInitiator(&definition, &workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified})

	for _, validation := range workflowcore.ValidateDefinition(definition) {
		if validation.Code == workflowcore.ValidationInitiator {
			return
		}
	}
	t.Fatal("empty specified initiator configuration should fail definition validation")
}

func TestValidateDesignDraftChecksFlowConfiguration(t *testing.T) {
	definition := newDefaultDefinition("purchase_approval", "采购申请审批")
	definition.Nodes[0].Initiator = &workflowcore.InitiatorConfig{Scope: workflowcore.InitiatorScopeSpecified}

	validations := validateDesignDraft(definition)
	if len(validations) != 1 || validations[0].Code != workflowcore.ValidationInitiator {
		t.Fatalf("design validation should check flow configuration, got %#v", validations)
	}
}
