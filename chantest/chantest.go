package chantest

import (
	"context"
	"time"

	"github.com/ezotaka/golib/conv"
)

// Type of context key
type ctxKey int

const (
	// Key of count to cancel channel
	countToCancelKey ctxKey = iota
)

// Get count to cancel channel
func countToCancel(ctx context.Context) (int, bool) {
	cnt, ok := ctx.Value(countToCancelKey).(int)
	return cnt, ok
}

// Get context with cancellation by count
//
// It panics if parent is nil or  cnt is negative
func ContextWithCountCancel(cnt int) context.Context {
	if cnt < 0 {
		panic("cnt must be zero or positive")
	}
	return context.WithValue(context.Background(), countToCancelKey, cnt)
}

func ContextWithTimeout(t time.Duration) context.Context {
	//ctx, _ := context.WithTimeout(context.Background(), t)
	//* above code is warned like below
	// the cancel function returned by context.WithTimeout should be called, not discarded, to avoid a context leak

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(t)
		cancel()
	}()
	return ctx
}

// TODO: args must be replaced with Case[C, A]
// TODO: catch panic and return error
func Run[C any, A any](
	ctx context.Context,
	fn func(context.Context, A) <-chan C,
	args A,
) []C {
	if fn == nil {
		panic("fn must not be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	c := fn(ctx, args)
	testChan := make(chan C)
	go func() {
		defer close(testChan)
		defer cancel()

		if cnt, ok := countToCancel(ctx); ok && cnt == 0 {
			return
		}

		count := 1
		for {
			select {
			case <-ctx.Done():
				return
			default:
				select {
				case val, ok := <-c:
					if !ok {
						return
					}
					testChan <- val
					if cnt, ok := countToCancel(ctx); ok {
						if cnt == count {
							return
						}
					}
					count++
				default:
				}
			}
		}
	}()
	return conv.ToSlice(context.Background(), testChan)
}

// TODO: Add type parameter W which is type of return of function to be tested
// Test case for function like func(done <-chan any, [spread Args]) <-chan C
type Case[C any, A any] struct {
	// Name of test case
	Name string

	// Args passed to the target method
	Args A

	// Context to cancel the channel which is return of function to be tested
	Context context.Context

	// Expected channel values
	Want []C
}
