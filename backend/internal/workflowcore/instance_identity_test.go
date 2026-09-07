package workflowcore

import (
	"testing"
	"time"
)

func TestResolveInstanceIdentityUsesMonthlyWindowStartAsBusinessMonth(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	definition := validLinearDefinition()
	definition.Name = "绩效考评单"
	definition.InstanceIdentity = &InstanceIdentityConfig{
		TitleTemplate: "{{businessPeriod}} {{starterName}}绩效考评",
		BusinessPeriod: &BusinessPeriodConfig{
			Enabled: true, Granularity: BusinessPeriodGranularityMonth,
			Source: BusinessPeriodSourceAvailabilityWindowStart,
		},
	}
	definition.Nodes[0].Availability = &StartAvailabilityConfig{
		Mode: StartAvailabilityMonthlyWindow, Timezone: "Asia/Shanghai",
		WindowStartDayFromEnd: 2, WindowStartTime: "09:00",
		WindowEndDay: 5, WindowEndTime: "18:00",
	}

	for _, submittedAt := range []time.Time{
		time.Date(2026, time.August, 31, 10, 0, 0, 0, location),
		time.Date(2026, time.September, 3, 10, 0, 0, 0, location),
	} {
		identity, err := ResolveInstanceIdentity(definition, submittedAt, "Foster", nil)
		if err != nil {
			t.Fatalf("ResolveInstanceIdentity(%s) error = %v", submittedAt, err)
		}
		if identity.BusinessPeriodKey != "2026-08" || identity.BusinessPeriodLabel != "2026年8月" {
			t.Fatalf("ResolveInstanceIdentity(%s) period = %#v", submittedAt, identity)
		}
		if identity.Title != "2026年8月 Foster绩效考评" {
			t.Fatalf("ResolveInstanceIdentity(%s) title = %q", submittedAt, identity.Title)
		}
	}
}

func TestResolveInstanceIdentitySupportsFormDateAndFieldVariables(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	definition := validLinearDefinition()
	definition.Name = "请假审批（华东区）"
	definition.DisplayName = "请假审批"
	definition.Form = []FormField{
		{Key: "reviewDate", Label: "考评日期", Type: FormFieldTypeDate},
		{Key: "project", Label: "项目", Type: FormFieldTypeText},
	}
	definition.InstanceIdentity = &InstanceIdentityConfig{
		TitleTemplate: "{{businessPeriod}} {{field.project}} - {{workflowName}}",
		BusinessPeriod: &BusinessPeriodConfig{
			Enabled: true, Granularity: BusinessPeriodGranularityQuarter,
			Source: BusinessPeriodSourceFormField, Field: "reviewDate", Offset: -1,
		},
	}

	identity, err := ResolveInstanceIdentity(
		definition,
		time.Date(2026, time.September, 5, 10, 0, 0, 0, location),
		"Foster",
		map[string]interface{}{"reviewDate": "2026-09-01", "project": "Phoenix"},
	)
	if err != nil {
		t.Fatalf("ResolveInstanceIdentity() error = %v", err)
	}
	if identity.BusinessPeriodKey != "2026-Q2" || identity.BusinessPeriodLabel != "2026年第2季度" {
		t.Fatalf("business period = %#v", identity)
	}
	if identity.Title != "2026年第2季度 Phoenix - 请假审批" {
		t.Fatalf("title = %q", identity.Title)
	}
}

func TestResolveInstanceIdentityDefaultTitleUsesDisplayName(t *testing.T) {
	definition := validLinearDefinition()
	definition.Name = "采购审批（集团总部）"
	definition.DisplayName = "采购申请"

	identity, err := ResolveInstanceIdentity(definition, time.Now(), "Foster", nil)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Title != "Foster提交的采购申请" {
		t.Fatalf("title = %q", identity.Title)
	}
}

func TestValidateDefinitionRejectsInvalidInstanceIdentity(t *testing.T) {
	tests := []struct {
		name   string
		config *InstanceIdentityConfig
	}{
		{name: "unsupported title variable", config: &InstanceIdentityConfig{TitleTemplate: "{{unknown}}"}},
		{name: "period variable without period", config: &InstanceIdentityConfig{TitleTemplate: "{{businessPeriod}}"}},
		{name: "missing form date field", config: &InstanceIdentityConfig{
			TitleTemplate: "{{workflowName}}", BusinessPeriod: &BusinessPeriodConfig{
				Enabled: true, Granularity: BusinessPeriodGranularityMonth,
				Source: BusinessPeriodSourceFormField, Field: "missing",
			},
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.InstanceIdentity = test.config
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationInstanceIdentity) {
				t.Fatalf("expected %s, got %#v", ValidationInstanceIdentity, errors)
			}
		})
	}
}

func TestValidateDefinitionRejectsWindowBusinessPeriodWithoutWindowAvailability(t *testing.T) {
	definition := validLinearDefinition()
	definition.InstanceIdentity = &InstanceIdentityConfig{
		TitleTemplate: "{{businessPeriod}} {{workflowName}}",
		BusinessPeriod: &BusinessPeriodConfig{
			Enabled: true, Granularity: BusinessPeriodGranularityMonth,
			Source: BusinessPeriodSourceAvailabilityWindowStart,
		},
	}
	definition.Nodes[0].Availability = &StartAvailabilityConfig{Mode: StartAvailabilityAlways}

	if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationInstanceIdentity) {
		t.Fatalf("expected %s, got %#v", ValidationInstanceIdentity, errors)
	}
}
