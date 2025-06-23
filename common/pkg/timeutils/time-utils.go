package timeutils

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAllAttemptsFailed = errors.New("all attempts failed")
)

func RetryRes[T any](
	ctx context.Context,
	attemptDelays []time.Duration,
	function func(context.Context) (T, error),
	onError func(context.Context, error) (needRetry bool),
	infinite bool,
) (T, error) {
	for i := 0; i < len(attemptDelays); {
		if ctx.Err() != nil {
			var res T
			return res, fmt.Errorf("retry canceled: %w", ctx.Err())
		}
		res, err := function(ctx)
		if err == nil {
			return res, err
		}
		needRetry := onError(ctx, err)
		if !needRetry {
			var zero T
			return zero, err
		}
		err = SleepCtx(ctx, attemptDelays[i])
		if err != nil {
			var zero T
			return zero, err
		}

		i++
		if infinite {
			i = min(i, len(attemptDelays)-1)
		}
	}

	var res T
	return res, ErrAllAttemptsFailed
}

func Retry(
	ctx context.Context,
	attemptDelays []time.Duration,
	function func(context.Context) error,
	onError func(context.Context, error) (needRetry bool),
	infinite bool,
) error {
	for i := 0; i < len(attemptDelays); {
		if ctx.Err() != nil {

			return fmt.Errorf("retry canceled: %w", ctx.Err())
		}
		err := function(ctx)
		if err == nil {
			return nil
		}
		if !onError(ctx, err) {
			return err
		}
		err = SleepCtx(ctx, attemptDelays[i])
		if err != nil {

			return err
		}

		i++
		if infinite {
			i = min(i, len(attemptDelays)-1)
		}
	}

	return ErrAllAttemptsFailed
}

func SleepCtx(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("sleep canceled: %w", ctx.Err())
	case <-time.After(d):
		return nil
	}
}
