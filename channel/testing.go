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

// TODO: rename method and comment
func DoTest[
	C any,
	A any,
](
	test TestCase[C, A],
	caller func(<-chan any, A) <-chan C,
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
