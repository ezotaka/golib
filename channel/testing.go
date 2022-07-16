package channel

import (
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
		<-chan any, // done channel
		A, // Args passed to the test target method
	) <-chan C, // Return of the function to be tested
) []C {
	done := make(chan any)
	returnChan := caller(done, test.Args)
	got := []C{}
	for val := range orTestCaseDone(done, &test, returnChan) {
		got = append(got, val)
	}
	return got
}

// Return channel which is closed when c is closed or conditions in test case are met
func orTestCaseDone[C any, A any](done chan any, t *TestCase[C, A], c <-chan C) <-chan C {
	doneClosed := make(chan any)
	returnChan := make(chan C)

	// * type declarations inside generic functions are not currently supported
	forChan := make(chan struct {
		index int
		value C
	})

	closeDone := func() {
		close(done)
		close(doneClosed)
	}

	if t.IsDoneAtFirst {
		closeDone()
	}

	// main goroutine which iterate c channel
	go func() {
		defer close(returnChan)
		defer close(forChan)
		index := 0
		for v := range c {
			select {
			case <-doneClosed:
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

	if t.IsDoneByTime > 0 {
		go func() {
			for {
				select {
				case <-doneClosed:
					return
				case <-time.After(t.IsDoneByTime):
					closeDone()
					return
				}
			}
		}()
	}

	go func() {
	loop:
		for {
			select {
			case <-doneClosed:
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
		closeDone()
	}()

	return returnChan
}
