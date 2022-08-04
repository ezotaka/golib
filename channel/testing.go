package channel

import (
	"context"

	"github.com/ezotaka/golib/conv"
	"github.com/ezotaka/golib/eztest"
)

// TODO: copy and paste from ctxpl package
// return channel which is closed when channel or done is closed
func orDone[T any](
	ctx context.Context,
	channel <-chan T,
) <-chan T {
	valChan := make(chan T)
	go func() {
		defer close(valChan)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-channel:
				if !ok {
					return
				}
				select {
				case valChan <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return valChan
}

func take[T any](
	ctx context.Context,
	valueChan <-chan T,
	num int,
) <-chan T {
	if valueChan == nil {
		return nil
	}
	takeChan := make(chan T)
	go func() {
		defer close(takeChan)
		for i := 0; i < num; i++ {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-valueChan:
				if !ok {
					return
				}
				takeChan <- v
			}
		}
	}()
	return takeChan
}

// return channel which can be cancelled by context
func withCountCancel[T any](ctx context.Context, c <-chan T) <-chan T {
	if cnt, ok := eztest.CountToCancel(ctx); ok {
		return orDone(ctx, take(ctx, c, cnt))
	} else {
		return orDone(ctx, c)
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
		return conv.Slice(withCountCancel(ctx, c)), nil
	}

	return eztest.Run(tc, pp)
}
