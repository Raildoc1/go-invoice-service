package chutils

import "context"

func LoopCtx(ctx context.Context, f func(context.Context)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			f(ctx)
		}
	}
}
