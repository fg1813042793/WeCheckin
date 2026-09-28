package workflowcore

import "testing"

func TestRuntimeStartConfigRoundTripKeepsFourWorkflowSettings(t *testing.T) {
	definition := validLinearDefinition()
	definition.InstanceIdentity = &InstanceIdentityConfig{
		TitleTemplate:  "{{businessPeriod}} {{workflowName}}",
		BusinessPeriod: &BusinessPeriodConfig{Enabled: true, Granularity: "month", Source: "submit_time"},
	}
	definition.Nodes[0].Initiator = &InitiatorConfig{Scope: InitiatorScopeSpecified, UserIDs: []uint{7}, ExcludedUserIDs: []uint{9}}
	definition.Nodes[0].Availability = &StartAvailabilityConfig{Mode: StartAvailabilityWeekly, Timezone: "Asia/Shanghai", Weekdays: []int{1, 3}, DailyStartTime: "09:00", DailyEndTime: "18:00"}
	definition.Nodes[0].StartLimit = &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodWeek, MaxCount: 2}

	raw, err := EncodeRuntimeStartConfig(definition)
	if err != nil {
		t.Fatalf("encode runtime start config: %v", err)
	}
	target := validLinearDefinition()
	if err := ApplyRuntimeStartConfigJSON(&target, raw); err != nil {
		t.Fatalf("apply runtime start config: %v", err)
	}
	if target.Nodes[0].Initiator.Scope != InitiatorScopeSpecified || target.Nodes[0].Initiator.UserIDs[0] != 7 {
		t.Fatalf("initiator = %#v", target.Nodes[0].Initiator)
	}
	if target.Nodes[0].Availability.Mode != StartAvailabilityWeekly || len(target.Nodes[0].Availability.Weekdays) != 2 {
		t.Fatalf("availability = %#v", target.Nodes[0].Availability)
	}
	if target.Nodes[0].StartLimit.MaxCount != 2 || target.InstanceIdentity.BusinessPeriod.Source != "submit_time" {
		t.Fatalf("limit/identity = %#v / %#v", target.Nodes[0].StartLimit, target.InstanceIdentity)
	}
}

func TestRuntimeStartConfigIsDefensivelyCloned(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[0].Initiator = &InitiatorConfig{Scope: InitiatorScopeSpecified, UserIDs: []uint{7}}
	config := RuntimeStartConfigFromDefinition(definition)
	config.Initiator.UserIDs[0] = 99
	if definition.Nodes[0].Initiator.UserIDs[0] != 7 {
		t.Fatal("runtime config extraction must not mutate the definition")
	}
}

func TestValidateRuntimeStartConfigOnlyReturnsLiveConfigErrors(t *testing.T) {
	definition := validLinearDefinition()
	definition.Nodes[0].Initiator = &InitiatorConfig{Scope: InitiatorScopeSpecified}
	definition.Form = []FormField{{Key: "broken", Label: "", Type: FormFieldTypeText}}
	errors := ValidateRuntimeStartConfig(definition)
	if len(errors) != 1 || errors[0].Code != ValidationInitiator {
		t.Fatalf("runtime start config errors = %#v", errors)
	}
}
