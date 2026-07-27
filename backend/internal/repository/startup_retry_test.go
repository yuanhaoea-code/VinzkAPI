package repository

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitForDependencyRetriesUntilReady(t *testing.T) {
	var attempts atomic.Int32
	err := waitForDependency(context.Background(), "test dependency", 2*time.Second, func(context.Context) error {
		if attempts.Add(1) < 2 {
			return errors.New("not ready")
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, int32(2), attempts.Load())
}

func TestWaitForDependencyHonorsTimeout(t *testing.T) {
	err := waitForDependency(context.Background(), "test dependency", 20*time.Millisecond, func(context.Context) error {
		return errors.New("not ready")
	})

	require.ErrorContains(t, err, "wait for test dependency readiness")
}
