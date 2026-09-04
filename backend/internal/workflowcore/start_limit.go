package workflowcore

import (
	"fmt"
	"strings"
	"time"
)

func DefaultStartLimit() StartLimitConfig {
	return StartLimitConfig{Mode: StartLimitModeUnlimited}
}

func CloneStartLimit(config *StartLimitConfig) StartLimitConfig {
	if config == nil || strings.TrimSpace(config.Mode) == "" {
		return DefaultStartLimit()
	}
	return StartLimitConfig{
		Mode:     strings.TrimSpace(config.Mode),
		Period:   strings.TrimSpace(config.Period),
		MaxCount: config.MaxCount,
	}
}

func ResolveStartLimitWindow(
	limit *StartLimitConfig,
	availability *StartAvailabilityConfig,
	now time.Time,
) (StartLimitWindow, bool) {
	if !validStartLimitConfig(limit, availability) {
		return StartLimitWindow{}, false
	}
	config := CloneStartLimit(limit)
	if config.Mode != StartLimitModeLimited {
		return StartLimitWindow{}, false
	}
	if config.Period == StartLimitPeriodTotal {
		return StartLimitWindow{PeriodKey: StartLimitPeriodTotal}, true
	}

	location, err := time.LoadLocation(startLimitTimezone(availability))
	if err != nil {
		return StartLimitWindow{}, false
	}
	localNow := now.In(location)
	switch config.Period {
	case StartLimitPeriodDay:
		return calendarStartLimitWindow(StartLimitPeriodDay, startOfDay(localNow), startOfDay(localNow).AddDate(0, 0, 1)), true
	case StartLimitPeriodWeek:
		start := startOfWeek(localNow)
		return calendarStartLimitWindow(StartLimitPeriodWeek, start, start.AddDate(0, 0, 7)), true
	case StartLimitPeriodMonth:
		start := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, location)
		return calendarStartLimitWindow(StartLimitPeriodMonth, start, start.AddDate(0, 1, 0)), true
	case StartLimitPeriodAvailability:
		return resolveAvailabilityStartLimitWindow(availability, localNow)
	default:
		return StartLimitWindow{}, false
	}
}

func validStartLimitConfig(limit *StartLimitConfig, availability *StartAvailabilityConfig) bool {
	if limit == nil {
		return true
	}
	mode := strings.TrimSpace(limit.Mode)
	if mode == "" || mode == StartLimitModeUnlimited {
		return strings.TrimSpace(limit.Period) == "" && limit.MaxCount == 0
	}
	if mode != StartLimitModeLimited || limit.MaxCount < 1 || limit.MaxCount > MaxStartLimitCount {
		return false
	}
	switch strings.TrimSpace(limit.Period) {
	case StartLimitPeriodTotal, StartLimitPeriodDay, StartLimitPeriodWeek, StartLimitPeriodMonth:
		return true
	case StartLimitPeriodAvailability:
		if availability == nil {
			return false
		}
		switch strings.TrimSpace(availability.Mode) {
		case StartAvailabilityFixed, StartAvailabilityWeekly, StartAvailabilityMonthly:
			return validStartAvailabilityConfig(availability)
		default:
			return false
		}
	default:
		return false
	}
}

func resolveAvailabilityStartLimitWindow(availability *StartAvailabilityConfig, localNow time.Time) (StartLimitWindow, bool) {
	if availability == nil {
		return StartLimitWindow{}, false
	}
	switch strings.TrimSpace(availability.Mode) {
	case StartAvailabilityFixed:
		return StartLimitWindow{
			PeriodKey: fmt.Sprintf("availability:fixed:%d", availability.StartsAt),
			StartsAt:  availability.StartsAt,
			EndsAt:    availability.EndsAt,
		}, true
	case StartAvailabilityWeekly:
		start := startOfWeek(localNow)
		return calendarStartLimitWindow("availability:weekly", start, start.AddDate(0, 0, 7)), true
	case StartAvailabilityMonthly:
		start := time.Date(localNow.Year(), localNow.Month(), 1, 0, 0, 0, 0, localNow.Location())
		return calendarStartLimitWindow("availability:monthly", start, start.AddDate(0, 1, 0)), true
	default:
		return StartLimitWindow{}, false
	}
}

func calendarStartLimitWindow(prefix string, start, end time.Time) StartLimitWindow {
	return StartLimitWindow{
		PeriodKey: fmt.Sprintf("%s:%d", prefix, start.UnixMilli()),
		StartsAt:  start.UnixMilli(),
		EndsAt:    end.UnixMilli(),
	}
}

func startLimitTimezone(availability *StartAvailabilityConfig) string {
	if availability == nil {
		return DefaultStartAvailabilityTimezone
	}
	return normalizedStartAvailabilityTimezone(availability.Timezone)
}

func startOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func startOfWeek(value time.Time) time.Time {
	daysSinceMonday := (int(value.Weekday()) + 6) % 7
	return startOfDay(value).AddDate(0, 0, -daysSinceMonday)
}
