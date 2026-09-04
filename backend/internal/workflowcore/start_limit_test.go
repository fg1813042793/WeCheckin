package workflowcore

import (
	"testing"
	"time"
)

func TestValidateDefinitionAcceptsStartLimitModes(t *testing.T) {
	tests := []struct {
		name         string
		limit        *StartLimitConfig
		availability *StartAvailabilityConfig
	}{
		{name: "unset"},
		{name: "unlimited", limit: &StartLimitConfig{Mode: StartLimitModeUnlimited}},
		{name: "total", limit: &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodTotal, MaxCount: 1}},
		{name: "daily", limit: &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodDay, MaxCount: 2}},
		{name: "weekly", limit: &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodWeek, MaxCount: 3}},
		{name: "monthly", limit: &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodMonth, MaxCount: 4}},
		{
			name:  "availability window",
			limit: &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodAvailability, MaxCount: 1},
			availability: &StartAvailabilityConfig{
				Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{1},
				DailyStartTime: "09:00", DailyEndTime: "18:00",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].StartLimit = test.limit
			definition.Nodes[0].Availability = test.availability
			if errors := ValidateDefinition(definition); hasValidationCode(errors, ValidationStartLimit) {
				t.Fatalf("expected valid start limit, got %#v", errors)
			}
		})
	}
}

func TestValidateDefinitionRejectsInvalidStartLimit(t *testing.T) {
	tests := []struct {
		name         string
		limit        StartLimitConfig
		availability *StartAvailabilityConfig
	}{
		{name: "unknown mode", limit: StartLimitConfig{Mode: "once", Period: StartLimitPeriodTotal, MaxCount: 1}},
		{name: "limited without period", limit: StartLimitConfig{Mode: StartLimitModeLimited, MaxCount: 1}},
		{name: "limited without count", limit: StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodMonth}},
		{name: "count too large", limit: StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodMonth, MaxCount: 10001}},
		{
			name:         "availability period without recurring availability",
			limit:        StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodAvailability, MaxCount: 1},
			availability: &StartAvailabilityConfig{Mode: StartAvailabilityAlways},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition := validLinearDefinition()
			definition.Nodes[0].StartLimit = &test.limit
			definition.Nodes[0].Availability = test.availability
			if errors := ValidateDefinition(definition); !hasValidationCode(errors, ValidationStartLimit) {
				t.Fatalf("expected %s, got %#v", ValidationStartLimit, errors)
			}
		})
	}
}

func TestResolveStartLimitWindowUsesConfiguredTimezone(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.September, 3, 10, 30, 0, 0, location)

	tests := []struct {
		period    string
		wantKey   string
		wantStart time.Time
		wantEnd   time.Time
	}{
		{StartLimitPeriodDay, "day:1788364800000", time.Date(2026, 9, 3, 0, 0, 0, 0, location), time.Date(2026, 9, 4, 0, 0, 0, 0, location)},
		{StartLimitPeriodWeek, "week:1788105600000", time.Date(2026, 8, 31, 0, 0, 0, 0, location), time.Date(2026, 9, 7, 0, 0, 0, 0, location)},
		{StartLimitPeriodMonth, "month:1788192000000", time.Date(2026, 9, 1, 0, 0, 0, 0, location), time.Date(2026, 10, 1, 0, 0, 0, 0, location)},
	}
	for _, test := range tests {
		t.Run(test.period, func(t *testing.T) {
			window, ok := ResolveStartLimitWindow(
				&StartLimitConfig{Mode: StartLimitModeLimited, Period: test.period, MaxCount: 1},
				&StartAvailabilityConfig{Mode: StartAvailabilityAlways, Timezone: "Asia/Shanghai"},
				now,
			)
			if !ok {
				t.Fatal("ResolveStartLimitWindow() ok = false")
			}
			if window.PeriodKey != test.wantKey || window.StartsAt != test.wantStart.UnixMilli() || window.EndsAt != test.wantEnd.UnixMilli() {
				t.Fatalf("window = %#v", window)
			}
		})
	}
}

func TestResolveStartLimitWindowUsesAvailabilityCycle(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.September, 3, 10, 30, 0, 0, location)
	limit := &StartLimitConfig{Mode: StartLimitModeLimited, Period: StartLimitPeriodAvailability, MaxCount: 1}

	fixedStart := now.Add(-time.Hour)
	fixedEnd := now.Add(time.Hour)
	window, ok := ResolveStartLimitWindow(limit, &StartAvailabilityConfig{
		Mode: StartAvailabilityFixed, Timezone: "Asia/Shanghai", StartsAt: fixedStart.UnixMilli(), EndsAt: fixedEnd.UnixMilli(),
	}, now)
	if !ok || window.PeriodKey != "availability:fixed:1788399000000" || window.StartsAt != fixedStart.UnixMilli() || window.EndsAt != fixedEnd.UnixMilli() {
		t.Fatalf("fixed window = %#v, ok = %v", window, ok)
	}

	window, ok = ResolveStartLimitWindow(limit, &StartAvailabilityConfig{
		Mode: StartAvailabilityMonthly, Timezone: "Asia/Shanghai", MonthDays: []int{3}, DailyStartTime: "09:00", DailyEndTime: "18:00",
	}, now)
	if !ok || window.PeriodKey != "availability:monthly:1788192000000" || window.EndsAt != time.Date(2026, 10, 1, 0, 0, 0, 0, location).UnixMilli() {
		t.Fatalf("monthly window = %#v, ok = %v", window, ok)
	}
}
