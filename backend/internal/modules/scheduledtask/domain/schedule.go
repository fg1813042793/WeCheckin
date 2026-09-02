package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

const maxMissedScheduleScan = 100000

type Schedule struct {
	inner     cron.Schedule
	location  *time.Location
	precision string
}

type PreviewOccurrence struct {
	UTCMillis int64  `json:"utcMillis"`
	LocalTime string `json:"localTime"`
}

type DueRun struct {
	ScheduledAt    time.Time
	CoalescedCount int
}

type DueResult struct {
	Runs         []DueRun
	SkippedCount int
	Next         time.Time
}

func ParseSchedule(precision, expression, timezone string, minimumSecondInterval time.Duration) (*Schedule, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTimezone, timezone)
	}

	var parser cron.Parser
	var expectedFields int
	switch precision {
	case CronPrecisionMinute:
		parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		expectedFields = 5
	case CronPrecisionSecond:
		parser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		expectedFields = 6
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidCronPrecision, precision)
	}

	if strings.HasPrefix(strings.TrimSpace(expression), "@") || len(strings.Fields(expression)) != expectedFields {
		return nil, fmt.Errorf("%w: expected %d fields", ErrInvalidCronExpression, expectedFields)
	}
	inner, err := parser.Parse("CRON_TZ=" + timezone + " " + expression)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCronExpression, err)
	}
	schedule := &Schedule{inner: inner, location: location, precision: precision}
	if precision == CronPrecisionSecond && minimumSecondInterval > 0 {
		if err := schedule.validateMinimumInterval(minimumSecondInterval); err != nil {
			return nil, err
		}
	}
	return schedule, nil
}

func (s *Schedule) Next(after time.Time) time.Time {
	return s.inner.Next(after).UTC()
}

func (s *Schedule) Preview(after time.Time, count int) []PreviewOccurrence {
	if count <= 0 {
		return nil
	}
	result := make([]PreviewOccurrence, 0, count)
	cursor := after
	for len(result) < count {
		next := s.Next(cursor)
		if next.IsZero() {
			break
		}
		result = append(result, PreviewOccurrence{
			UTCMillis: next.UnixMilli(),
			LocalTime: next.In(s.location).Format("2006-01-02 15:04:05") + " " + s.location.String(),
		})
		cursor = next
	}
	return result
}

func (s *Schedule) ComputeDue(cursor, now time.Time, policy string, maxCatchUp int) (DueResult, error) {
	if err := ValidateMisfirePolicy(policy); err != nil {
		return DueResult{}, err
	}
	if maxCatchUp < 0 {
		maxCatchUp = 0
	}

	due := make([]time.Time, 0)
	next := s.Next(cursor)
	for !next.IsZero() && !next.After(now) {
		if len(due) >= maxMissedScheduleScan {
			return DueResult{}, ErrTooManyMissedSchedules
		}
		due = append(due, next)
		next = s.Next(next)
	}
	result := DueResult{Next: next}
	if len(due) == 0 {
		return result, nil
	}

	switch policy {
	case MisfirePolicySkip:
		result.SkippedCount = len(due)
	case MisfirePolicyFireOnce:
		result.Runs = []DueRun{{ScheduledAt: due[len(due)-1], CoalescedCount: len(due)}}
	case MisfirePolicyCatchUp:
		limit := maxCatchUp
		if limit > len(due) {
			limit = len(due)
		}
		result.Runs = make([]DueRun, 0, limit)
		for _, scheduledAt := range due[:limit] {
			result.Runs = append(result.Runs, DueRun{ScheduledAt: scheduledAt, CoalescedCount: 1})
		}
		result.SkippedCount = len(due) - limit
	}
	return result, nil
}

func (s *Schedule) validateMinimumInterval(minimum time.Duration) error {
	cursor := time.Now().UTC().Add(-time.Second)
	previous := s.Next(cursor)
	if previous.IsZero() {
		return nil
	}
	for i := 0; i < 512; i++ {
		next := s.Next(previous)
		if next.IsZero() {
			return nil
		}
		if next.Sub(previous) < minimum {
			return fmt.Errorf("%w: got %s, minimum %s", ErrScheduleTooFrequent, next.Sub(previous), minimum)
		}
		previous = next
	}
	return nil
}
