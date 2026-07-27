package repository

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	startupRetryInitialBackoff = 250 * time.Millisecond
	startupRetryMaxBackoff     = 5 * time.Second
	startupDependencyCheckWait = 5 * time.Second
)

func waitForDependency(ctx context.Context, name string, maxWait time.Duration, check func(context.Context) error) error {
	if check == nil {
		return fmt.Errorf("wait for %s: nil readiness check", name)
	}
	if maxWait <= 0 {
		checkCtx, cancel := context.WithTimeout(ctx, startupDependencyCheckWait)
		defer cancel()
		if err := check(checkCtx); err != nil {
			return fmt.Errorf("wait for %s: %w", name, err)
		}
		return nil
	}

	waitCtx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()
	backoff := startupRetryInitialBackoff
	var lastErr error
	for attempt := 1; ; attempt++ {
		checkCtx, checkCancel := context.WithTimeout(waitCtx, startupDependencyCheckWait)
		err := check(checkCtx)
		checkCancel()
		if err == nil {
			if attempt > 1 {
				log.Printf("%s became ready after %d attempts", name, attempt)
			}
			return nil
		}
		lastErr = err
		log.Printf("%s is not ready (attempt %d): %v; retrying in %s", name, attempt, err, backoff)
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("wait for %s readiness: %w", name, lastErr)
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > startupRetryMaxBackoff {
			backoff = startupRetryMaxBackoff
		}
	}
}
