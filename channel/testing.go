package channel

import (
	"context"

	"github.com/ezotaka/golib/conv"
	"github.com/ezotaka/golib/eztest"
)

// return channel which can be cancelled by context
func withCountCancel[T any](ctx context.Context, c <-chan T) <-chan T {
	if cnt, ok := eztest.CountToCancel(ctx); ok {
		return OrDone(ctx, Take(ctx, c, cnt))
	} else {
		return OrDone(ctx, c)
	}
}

// RunTest channel test using test case defined by Case
func RunTest[
	// Type of args which is passed to the function to be tested
	A any,
	// Type of chanel which is returned by the function to be tested
	C any,
](
	// Test case for the function to be tested
	tc eztest.Case[A, <-chan C, []C],
) ([]C, error) {
	// [post process]
	// Channel c which is returned by the function to be tested can be canceled by the context.
	// Synchronously converts the value sent from the channel into slices.
	pp := func(ctx context.Context, c <-chan C, err error) ([]C, error) {
		if c == nil || err != nil {
			return nil, err
		}
		return conv.ToSlice(context.Background(), withCountCancel(ctx, c)), nil
	}

	return eztest.Run(tc, pp)
}
