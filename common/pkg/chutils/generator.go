package chutils

import (
	"context"
	"time"
)

func Generator[T any](
	ctx context.Context,
	buffCap int32,
	minInterval time.Duration,
	tick func(ctx context.Context, buffLen int32) ([]T, error),
) (<-chan T, <-chan error) {
	out := make(chan T, buffCap)
	errCh := make(chan error)

	go func(ctx context.Context) {
		defer close(out)
		defer close(errCh)

		var nextRequest time.Time

		LoopCtx(
			ctx,
			func(ctx context.Context) {
				nextRequest = time.Now().Add(minInterval)
				val, err := tick(ctx, int32(len(out)))
				if err != nil {
					errCh <- err
					return
				}
				for _, msg := range val {
					out <- msg
				}
				timeToWait := nextRequest.Sub(time.Now())
				if timeToWait > 0 {
					time.Sleep(timeToWait)
				}
			},
		)
	}(ctx)

	return out, errCh
}
