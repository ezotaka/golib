// Copyright (c) 2022 Takatomo Ezo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eztest

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// Type of context key
type ctxKey int

const (
	// Key of count to cancel channel
	countToCancelKey ctxKey = iota
)

// TODO: replace with receiver
// Get count to cancel channel
func CountToCancel(ctx context.Context) (int, bool) {
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

// Type of invoker the method to be tested
type Invoker[A any, R any] struct {
	Name   string
	Invoke func(context.Context, A) (
		R, // value returned by the function to be invoked
		error, // error returned by the function to be invoked
	)
}

// Test case for function like func(context.Context, [spread A]) <-chan C
type Case[
	// Args type of the function to be tested
	A any,
	// Return type of the function to be tested
	R any,
	// Want type of the function test
	W any,
] struct {
	// Name of test case
	Name string

	// Context to cancel the channel which is return of function to be tested
	Context context.Context

	// Args passed to the target method
	Args A

	// ? use Call (https://pkg.go.dev/reflect#Value.Call)
	// Invoker the method to be tested
	Invoker Invoker[A, R]

	// Expected channel values
	Want W

	// Expected error
	ErrMsg string

	// Expected panic
	Panic any
}

func notPanicMsg(name string, want any) string {
	return fmt.Sprintf("%s() doesn't panic, want panic '%v'", name, want)
}

func wrongPanicMsg(name string, got, want any) string {
	return fmt.Sprintf("%s() panic '%v', want panic '%v'", name, got, want)
}

func notErrorMsg(name, want string) string {
	return fmt.Sprintf("%s() doesn't error, want error '%v'", name, want)
}

func wrongErrorMsg(name, got, want string) string {
	return fmt.Sprintf("%s() error '%v', want error '%v'", name, got, want)
}

func notEqualsMsg[W any](name string, got, want W) string {
	return fmt.Sprintf("%s() = %v, want %v", name, got, want)
}

// Run test using test case defined by Case
func Run[
	// Args type of the function to be tested
	A any,
	// Return type of the function to be tested
	R any,
	// Return type of Run function
	W any,
](
	// Test case for the function to be tested
	tc Case[A, R, W],
	// PostProcessor return of the function to be tested
	pp func(context.Context, R, error) (W, error),
) (got W, err error) {
	panicInRun := false // ? not needed ?
	if tc.Invoker.Invoke == nil {
		panicInRun = true
		panic("c.Invoker.Invoke must not be nil")
	}

	name := tc.Invoker.Name
	var errMsg string
	var panicVal any

	defer func() {
		if !panicInRun {
			if r := recover(); r != nil {
				panicVal = r
			}
		}

		if panicVal != nil {
			if panicVal != tc.Panic {
				err = fmt.Errorf(wrongPanicMsg(name, panicVal, tc.Panic))
			}
			return
		} else if tc.Panic != nil {
			err = fmt.Errorf(notPanicMsg(name, tc.Panic))
			return
		}

		if errMsg != "" {
			if errMsg != tc.ErrMsg {
				err = fmt.Errorf(wrongErrorMsg(name, errMsg, tc.ErrMsg))
			}
			return
		} else if tc.ErrMsg != "" {
			err = fmt.Errorf(notErrorMsg(name, tc.ErrMsg))
			return
		}

		if !reflect.DeepEqual(got, tc.Want) {
			err = fmt.Errorf(notEqualsMsg(name, got, tc.Want))
		}
	}()

	ctx := tc.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// invoke the method to be tested
	testedVal, testedErr := tc.Invoker.Invoke(ctx, tc.Args)

	// post process
	ppVal, ppErr := pp(ctx, testedVal, testedErr)

	if ppErr == nil {
		got = ppVal
	} else {
		errMsg = ppErr.Error()
	}
	return
}
