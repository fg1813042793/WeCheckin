package domain

import (
	"errors"
	"fmt"
)

const (
	CronPrecisionMinute = "minute"
	CronPrecisionSecond = "second"

	MisfirePolicySkip     = "skip"
	MisfirePolicyFireOnce = "fire_once"
	MisfirePolicyCatchUp  = "catch_up"

	ConcurrencyPolicySkip      = "skip"
	ConcurrencyPolicyQueueOnce = "queue_once"
	ConcurrencyPolicyAllow     = "allow"
)

var (
	ErrInvalidCronPrecision     = errors.New("invalid cron precision")
	ErrInvalidCronExpression    = errors.New("invalid cron expression")
	ErrInvalidTimezone          = errors.New("invalid timezone")
	ErrScheduleTooFrequent      = errors.New("schedule interval is below the configured minimum")
	ErrInvalidMisfirePolicy     = errors.New("invalid misfire policy")
	ErrInvalidConcurrencyPolicy = errors.New("invalid concurrency policy")
	ErrTooManyMissedSchedules   = errors.New("too many missed schedules")
)

func ValidateMisfirePolicy(policy string) error {
	switch policy {
	case MisfirePolicySkip, MisfirePolicyFireOnce, MisfirePolicyCatchUp:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidMisfirePolicy, policy)
	}
}

func ValidateConcurrencyPolicy(policy string) error {
	switch policy {
	case ConcurrencyPolicySkip, ConcurrencyPolicyQueueOnce, ConcurrencyPolicyAllow:
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrInvalidConcurrencyPolicy, policy)
	}
}
