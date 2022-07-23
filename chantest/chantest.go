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

// TODO: Add type parameter R which is type of return of function to be tested
// Test case for function like func(context.Context, [spread A]) <-chan C
type Case[C any, A any] struct {
	// Name of test case
	Name string

	// Args passed to the target method
	Args A

	// Context to cancel the channel which is return of function to be tested
	Context context.Context

	// Invoker the method to be tested
	Invoker func(context.Context, A) (<-chan C, error)

	// Expected channel values
	Want []C

	// Expected error
	ErrMsg string

	// Expected panic
	Panic any
}

type PanicError error

// TODO: Run function precesses assertion
// Run test using test case defined by Case
func Run[
	// Type of chanel which is returned by the function to be tested
	C any,
	// Type of args which is passed to the function to be tested
	A any,
](
	// Test case for the function to be tested
	tc Case[C, A],
) (
	// Channel which is returned by the function to be tested
	ret []C,
	// Message of error which is returned by the function to be tested
	errMsg string,
	// Value of panic which is caused by the function to be tested
	panicVal any,
) {
	panicInRun := false
	defer func() {
		if !panicInRun {
			if r := recover(); r != nil {
				panicVal = r
			}
		}
	}()

	if tc.Invoker == nil {
		panicInRun = true
		panic("c.Invoker must not be nil")
	}

	ctx := tc.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// invoke the method to be tested
	returnedChan, returnedErr := tc.Invoker(ctx, tc.Args)
	if returnedErr != nil {
		errMsg = returnedErr.Error()
		return
	}

	endedChan := make(chan C)
	go func() {
		defer close(endedChan)

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
				case val, ok := <-returnedChan:
					if !ok {
						return
					}
					endedChan <- val
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
	//return conv.ToSlice(context.Background(), endedChan), err
	ret = conv.ToSlice(context.Background(), endedChan)
	return
}
