package testing

import (
	"time"
)

// Test case for function like func(done <-chan any, [spread Args]) <-chan C
type ChanFuncTestCase[C any, A any] struct {
	// Name of test case
	Name string

	// Args passed to the target method
	Args A

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

func ExecChanFuncTest[
	C any,
	A any,
](
	test ChanFuncTestCase[C, A],
	call func(<-chan any, A) <-chan C,
) []C {
	done := make(chan any)
	returnChan := call(done, test.Args)
	got := []C{}
	for val := range orTestCaseDone(done, &test, returnChan) {
		got = append(got, val)
	}
	return got
}

// Return channel which is closed when c is closed or conditions in test case are met
func orTestCaseDone[C any, A any](done chan any, t *ChanFuncTestCase[C, A], c <-chan C) <-chan C {
	returnChan := make(chan C)
	indexChan := make(chan int)
	valChan := make(chan C)

	go func() {
		defer close(returnChan)
		index := 0
		for v := range c {
			returnChan <- v
			indexChan <- index
			valChan <- v
		}
	}()

	go func() {
	loop:
		for {
			select {
			case <-time.After(t.IsDoneByTime):
				if t.IsDoneByTime > 0 {
					break loop
				}
			case i := <-indexChan:
				if t.IsDoneByIndex != nil && t.IsDoneByIndex(i) {
					break loop
				}
			case v := <-valChan:
				if t.IsDoneByValue != nil && t.IsDoneByValue(v) {
					break loop
				}
			}
		}
		close(done)
	}()

	return returnChan
}
