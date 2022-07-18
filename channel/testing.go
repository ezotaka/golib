package channel

import (
	"context"
	"sync"
	"time"
)

// Test case for function like func(done <-chan any, [spread Args]) <-chan C
type TestCase[C any, A any] struct {
	// Name of test case
	Name string

	// Args passed to the target method
	Args A

	// "done" channel is already closed when called the target method
	IsDoneAtFirst bool

	// Index condition for closing the "done" channel
	// This is called on every iteration
	IsDoneByIndex func(
		int, // now index number
	) bool // to close "done" channel or not

	// Value condition for closing the "done" channel
	// This is called on every iteration
	IsDoneByValue func(
		C, // now channel value
	) bool // to close "done" channel or not

	// Value condition for closing the "done" channel
	// This is called on every iteration
	IsDoneByTime time.Duration // since test case started

	// Expected channel values
	Want []C
}

// Execute the function to be tested using caller,
// and read returned channel to end,
// then return read values as []C.
// The function to be tested like func(done <- chan any, [spread test.Args]) <- chan C
// done channel is closed when conditions in TestCase are met,
func GetTestedValues[
	// Type of returned channel
	C any,
	// Type of args passed to the test target method
	A any,
](
	// Test case for the function to be tested
	test TestCase[C, A],
	// caller executes the function to be tested
	caller func(
		context.Context, // context
		A, // Args passed to the test target method
	) <-chan C, // Return of the function to be tested
) []C {
	ctx, cancel := context.WithCancel(context.Background())
	if test.IsDoneAtFirst {
		// already past the deadline
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-1*time.Second))
	}
	if test.IsDoneByTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, test.IsDoneByTime)
	}

	returnChan := caller(ctx, test.Args)
	got := []C{}
	for val := range orTestCaseDone(ctx, cancel, &test, returnChan) {
		got = append(got, val)
	}
	return got
}

// Return channel which is closed when c is closed or conditions in test case are met
func orTestCaseDone[C any, A any](ctx context.Context, cancel context.CancelFunc, t *TestCase[C, A], c <-chan C) <-chan C {
	returnChan := make(chan C)

	// * type declarations inside generic functions are not currently supported
	forChan := make(chan struct {
		index int
		value C
	})

	// main goroutine which iterate c channel
	go func() {
		defer close(returnChan)
		defer close(forChan)
		index := 0
		for v := range c {
			select {
			case <-ctx.Done():
				// wait a bit until c channel maybe closed
				time.Sleep(10 * time.Millisecond)
			case returnChan <- v:
				forChan <- struct {
					index int
					value C
				}{
					index: index,
					value: v,
				}
				index++
			}
		}
	}()

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			case f := <-forChan:
				if t.IsDoneByIndex != nil && t.IsDoneByIndex(f.index) {
					break loop
				}
				if t.IsDoneByValue != nil && t.IsDoneByValue(f.value) {
					break loop
				}
			}
		}
		cancel()
	}()

	return returnChan
}

// Execute the function to be tested using caller,
// and read returned channel to end,
// then return read values as []C.
// The function to be tested like func(done <- chan any, [spread test.Args]) <- chan C
// done channel is closed when conditions in TestCase are met,
func GetTestedValues2[
	// Type of returned channel
	C any,
	// Type of args passed to the test target method
	A any,
](
	// Test case for the function to be tested
	test TestCase[C, A],
	// caller executes the function to be tested
	caller func(
		context.Context, // context
		A, // Args passed to the test target method
	) (<-chan C, <-chan C), // Return of the function to be tested
) ([]C, []C) {
	ctx, cancel := context.WithCancel(context.Background())
	if test.IsDoneAtFirst {
		// already past the deadline
		ctx, cancel = context.WithDeadline(ctx, time.Now().Add(-1*time.Second))
	}
	if test.IsDoneByTime > 0 {
		ctx, cancel = context.WithTimeout(ctx, test.IsDoneByTime)
	}

	ch1, ch2 := caller(ctx, test.Args)
	getter := func(returnChan <-chan C) []C {
		got := []C{}
		for val := range orTestCaseDone(ctx, cancel, &test, returnChan) {
			got = append(got, val)
		}
		return got
	}

	var ret1, ret2 []C
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ret1 = getter(ch1)
	}()
	go func() {
		defer wg.Done()
		ret2 = getter(ch2)
	}()
	wg.Wait()
	return ret1, ret2
}
