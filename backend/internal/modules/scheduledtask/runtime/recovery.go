package runtime

import (
	"context"
	"log"
	"math/rand"
	"time"
)

const (
	defaultRetryInitialDelay = time.Second
	defaultRetryMaxDelay     = 30 * time.Second
)

type logFunc func(string, ...interface{})

type retryController struct {
	initial  time.Duration
	maximum  time.Duration
	current  time.Duration
	failures int
	logf     logFunc
}

func newRetryController(initial, maximum time.Duration, logf logFunc) *retryController {
	if initial <= 0 {
		initial = defaultRetryInitialDelay
	}
	if maximum < initial {
		maximum = defaultRetryMaxDelay
		if maximum < initial {
			maximum = initial
		}
	}
	if logf == nil {
		logf = log.Printf
	}
	return &retryController{initial: initial, maximum: maximum, logf: logf}
}

func (controller *retryController) WaitAfterFailure(ctx context.Context, component, operation string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if controller.current <= 0 {
		controller.current = controller.initial
	} else if controller.current < controller.maximum {
		controller.current *= 2
		if controller.current > controller.maximum {
			controller.current = controller.maximum
		}
	}
	controller.failures++
	delay := jitteredDelay(controller.current)
	controller.logf("scheduled task %s temporarily unavailable operation=%s retryIn=%s error=%v", component, operation, delay, err)
	return waitContext(ctx, delay)
}

func (controller *retryController) MarkSuccess(component string) {
	if controller.failures > 0 {
		controller.logf("scheduled task %s recovered after %d failure(s)", component, controller.failures)
	}
	controller.current = 0
	controller.failures = 0
}

func retryInitialDelay(interval time.Duration) time.Duration {
	if interval > 0 && interval < defaultRetryInitialDelay {
		return interval
	}
	return defaultRetryInitialDelay
}

func jitteredDelay(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	minimum := base * 4 / 5
	spread := base - minimum
	if spread <= 0 {
		return base
	}
	return minimum + time.Duration(rand.Int63n(int64(spread)+1))
}

func waitContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
