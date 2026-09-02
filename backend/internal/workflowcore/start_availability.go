package workflowcore

import (
	"strings"
	"time"
)

func EvaluateStartAvailability(config *StartAvailabilityConfig, now time.Time) string {
	if config == nil {
		return StartAvailabilityStateAvailable
	}
	mode := strings.TrimSpace(config.Mode)
	if mode == "" || mode == StartAvailabilityAlways {
		return StartAvailabilityStateAvailable
	}
	if !validStartAvailabilityConfig(config) {
		return StartAvailabilityStateOutsideWindow
	}
	if mode == StartAvailabilityFixed {
		nowMillis := now.UnixMilli()
		if nowMillis < config.StartsAt {
			return StartAvailabilityStateNotStarted
		}
		if nowMillis >= config.EndsAt {
			return StartAvailabilityStateExpired
		}
		return StartAvailabilityStateAvailable
	}

	location, _ := time.LoadLocation(normalizedStartAvailabilityTimezone(config.Timezone))
	localNow := now.In(location)
	date := localNow.Format("2006-01-02")
	if config.EffectiveStartDate != "" && date < config.EffectiveStartDate {
		return StartAvailabilityStateNotStarted
	}
	if config.EffectiveEndDate != "" && date > config.EffectiveEndDate {
		return StartAvailabilityStateExpired
	}
	if !startAvailabilityMatchesDay(config, localNow) {
		return StartAvailabilityStateOutsideWindow
	}
	minute := localNow.Hour()*60 + localNow.Minute()
	startMinute, _ := parseStartAvailabilityClock(config.DailyStartTime)
	endMinute, _ := parseStartAvailabilityClock(config.DailyEndTime)
	if minute < startMinute || minute >= endMinute {
		return StartAvailabilityStateOutsideWindow
	}
	return StartAvailabilityStateAvailable
}

func DefaultStartAvailability() StartAvailabilityConfig {
	return StartAvailabilityConfig{Mode: StartAvailabilityAlways, Timezone: DefaultStartAvailabilityTimezone}
}

func CloneStartAvailability(config *StartAvailabilityConfig) StartAvailabilityConfig {
	if config == nil {
		return DefaultStartAvailability()
	}
	cloned := *config
	if strings.TrimSpace(cloned.Mode) == "" {
		cloned.Mode = StartAvailabilityAlways
	}
	cloned.Timezone = normalizedStartAvailabilityTimezone(cloned.Timezone)
	cloned.Weekdays = append([]int(nil), config.Weekdays...)
	cloned.MonthDays = append([]int(nil), config.MonthDays...)
	return cloned
}

func validStartAvailabilityConfig(config *StartAvailabilityConfig) bool {
	if config == nil {
		return true
	}
	mode := strings.TrimSpace(config.Mode)
	if mode == "" || mode == StartAvailabilityAlways {
		return true
	}
	if _, err := time.LoadLocation(normalizedStartAvailabilityTimezone(config.Timezone)); err != nil {
		return false
	}
	switch mode {
	case StartAvailabilityFixed:
		return config.StartsAt > 0 && config.EndsAt > config.StartsAt
	case StartAvailabilityWeekly:
		return validStartAvailabilityDays(config.Weekdays, 1, 7) && validRecurringStartAvailability(config)
	case StartAvailabilityMonthly:
		return (config.LastDayOfMonth || len(config.MonthDays) > 0) &&
			(len(config.MonthDays) == 0 || validStartAvailabilityDays(config.MonthDays, 1, 31)) &&
			validRecurringStartAvailability(config)
	default:
		return false
	}
}

func validRecurringStartAvailability(config *StartAvailabilityConfig) bool {
	startMinute, startOK := parseStartAvailabilityClock(config.DailyStartTime)
	endMinute, endOK := parseStartAvailabilityClock(config.DailyEndTime)
	if !startOK || !endOK || endMinute <= startMinute {
		return false
	}
	startDate, startOK := parseStartAvailabilityDate(config.EffectiveStartDate)
	endDate, endOK := parseStartAvailabilityDate(config.EffectiveEndDate)
	return startOK && endOK && (startDate.IsZero() || endDate.IsZero() || !startDate.After(endDate))
}

func validStartAvailabilityDays(days []int, minimum, maximum int) bool {
	if len(days) == 0 {
		return false
	}
	seen := make(map[int]struct{}, len(days))
	for _, day := range days {
		if day < minimum || day > maximum {
			return false
		}
		if _, exists := seen[day]; exists {
			return false
		}
		seen[day] = struct{}{}
	}
	return true
}

func parseStartAvailabilityClock(value string) (int, bool) {
	parsed, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, false
	}
	return parsed.Hour()*60 + parsed.Minute(), true
}

func parseStartAvailabilityDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	return parsed, err == nil
}

func normalizedStartAvailabilityTimezone(value string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return DefaultStartAvailabilityTimezone
}

func startAvailabilityMatchesDay(config *StartAvailabilityConfig, now time.Time) bool {
	switch strings.TrimSpace(config.Mode) {
	case StartAvailabilityWeekly:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return containsStartAvailabilityDay(config.Weekdays, weekday)
	case StartAvailabilityMonthly:
		if containsStartAvailabilityDay(config.MonthDays, now.Day()) {
			return true
		}
		lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
		return config.LastDayOfMonth && now.Day() == lastDay
	default:
		return false
	}
}

func containsStartAvailabilityDay(days []int, expected int) bool {
	for _, day := range days {
		if day == expected {
			return true
		}
	}
	return false
}
