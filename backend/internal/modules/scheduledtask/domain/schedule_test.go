package domain

import (
	"errors"
	"testing"
	"time"
)

func TestParseScheduleSupportsMinuteAndSecondPrecision(t *testing.T) {
	minute, err := ParseSchedule("minute", "*/5 * * * *", "Asia/Shanghai", 5*time.Second)
	if err != nil {
		t.Fatalf("minute schedule: %v", err)
	}
	second, err := ParseSchedule("second", "*/10 * * * * *", "Asia/Shanghai", 5*time.Second)
	if err != nil {
		t.Fatalf("second schedule: %v", err)
	}

	after := time.Date(2026, time.September, 1, 0, 0, 1, 0, time.UTC)
	if got := minute.Next(after); !got.Equal(time.Date(2026, time.September, 1, 0, 5, 0, 0, time.UTC)) {
		t.Fatalf("minute next = %v", got)
	}
	if got := second.Next(after); !got.Equal(time.Date(2026, time.September, 1, 0, 0, 10, 0, time.UTC)) {
		t.Fatalf("second next = %v", got)
	}
}

func TestSecondScheduleRejectsIntervalBelowConfiguredMinimum(t *testing.T) {
	_, err := ParseSchedule("second", "*/2 * * * * *", "UTC", 5*time.Second)
	if !errors.Is(err, ErrScheduleTooFrequent) {
		t.Fatalf("error = %v, want ErrScheduleTooFrequent", err)
	}
}

func TestParseScheduleRejectsInvalidPrecisionTimezoneAndDescriptors(t *testing.T) {
	for _, test := range []struct {
		name       string
		precision  string
		expression string
		timezone   string
		want       error
	}{
		{name: "precision", precision: "hour", expression: "0 * * * *", timezone: "UTC", want: ErrInvalidCronPrecision},
		{name: "timezone", precision: "minute", expression: "0 * * * *", timezone: "Mars/Base", want: ErrInvalidTimezone},
		{name: "descriptor", precision: "minute", expression: "@hourly", timezone: "UTC", want: ErrInvalidCronExpression},
		{name: "minute field count", precision: "minute", expression: "0 0 * * * *", timezone: "UTC", want: ErrInvalidCronExpression},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseSchedule(test.precision, test.expression, test.timezone, 5*time.Second)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSchedulePreviewReturnsUTCAndTaskTimezoneValues(t *testing.T) {
	schedule, err := ParseSchedule("minute", "0 9 * * *", "Asia/Shanghai", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	preview := schedule.Preview(time.Date(2026, time.September, 1, 0, 30, 0, 0, time.UTC), 2)
	if len(preview) != 2 {
		t.Fatalf("preview count = %d", len(preview))
	}
	if preview[0].UTCMillis != time.Date(2026, time.September, 1, 1, 0, 0, 0, time.UTC).UnixMilli() {
		t.Fatalf("first UTC millis = %d", preview[0].UTCMillis)
	}
	if preview[0].LocalTime != "2026-09-01 09:00:00 Asia/Shanghai" {
		t.Fatalf("first local time = %q", preview[0].LocalTime)
	}
}

func TestComputeDueAppliesMisfirePolicies(t *testing.T) {
	schedule, err := ParseSchedule("minute", "* * * * *", "UTC", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	cursor := time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)
	now := cursor.Add(4*time.Minute + 30*time.Second)

	for _, test := range []struct {
		policy        string
		maxCatchUp    int
		wantRuns      int
		wantSkipped   int
		wantCoalesced int
	}{
		{policy: MisfirePolicySkip, wantRuns: 0, wantSkipped: 4},
		{policy: MisfirePolicyFireOnce, wantRuns: 1, wantCoalesced: 4},
		{policy: MisfirePolicyCatchUp, maxCatchUp: 2, wantRuns: 2, wantSkipped: 2},
	} {
		t.Run(test.policy, func(t *testing.T) {
			result, err := schedule.ComputeDue(cursor, now, test.policy, test.maxCatchUp)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Runs) != test.wantRuns || result.SkippedCount != test.wantSkipped {
				t.Fatalf("runs/skipped = %d/%d, want %d/%d", len(result.Runs), result.SkippedCount, test.wantRuns, test.wantSkipped)
			}
			if test.wantCoalesced > 0 && result.Runs[0].CoalescedCount != test.wantCoalesced {
				t.Fatalf("coalesced = %d, want %d", result.Runs[0].CoalescedCount, test.wantCoalesced)
			}
			if !result.Next.Equal(cursor.Add(5 * time.Minute)) {
				t.Fatalf("next = %v", result.Next)
			}
		})
	}
}

func TestValidatePoliciesAcceptsOnlySupportedValues(t *testing.T) {
	for _, policy := range []string{MisfirePolicySkip, MisfirePolicyFireOnce, MisfirePolicyCatchUp} {
		if err := ValidateMisfirePolicy(policy); err != nil {
			t.Fatalf("misfire policy %q: %v", policy, err)
		}
	}
	if !errors.Is(ValidateMisfirePolicy("all"), ErrInvalidMisfirePolicy) {
		t.Fatal("unsupported misfire policy must fail")
	}
	for _, policy := range []string{ConcurrencyPolicySkip, ConcurrencyPolicyQueueOnce, ConcurrencyPolicyAllow} {
		if err := ValidateConcurrencyPolicy(policy); err != nil {
			t.Fatalf("concurrency policy %q: %v", policy, err)
		}
	}
	if !errors.Is(ValidateConcurrencyPolicy("parallel"), ErrInvalidConcurrencyPolicy) {
		t.Fatal("unsupported concurrency policy must fail")
	}
}
