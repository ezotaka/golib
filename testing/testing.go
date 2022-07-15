package testing

import (
	"reflect"
	"testing"
	"time"
)

// Test case for function like func(done <-chan any, [spread Args]) <-chan C
type ChanFuncTestCase[C any, A any] struct {
	// Name of test case
	Name string

	// Args passed to the target method
	Args A

	// Conditions for closing the "done" channel
	// This is called on every iteration
	IsDone func(
		int, // now index number
		C, // now channel value
	) bool // to close "done" channel or not

	// Expected channel values
	Want []C
}

// Execute Table Driven Test for function which returns <-chan C
func ExecReadOnlyChanFuncTest[
	C any,
	A any,
	T ChanFuncTestCase[C, A],
](
	t *testing.T,
	tests []T,
	call func(<-chan any, A) (string, <-chan C),
) {
	for _, tt := range tests {
		ttt := ChanFuncTestCase[C, A](tt)
		t.Run(ttt.Name, func(t *testing.T) {
			done := make(chan any)

			name, gotChan := call(done, ttt.Args)

			// close done channel while iterating gotChan
			index := 0
			closed := false
			closer := func(val C) {
				if closed || ttt.IsDone == nil {
					return
				}
				if ttt.IsDone(index, val) {
					close(done)
					closed = true

					// wait for goroutine in gotChan finished
					// not definitive but sufficient
					time.Sleep(time.Second) // TODO: too long?
				}
				index++
			}

			got := []C{}
			for val := range gotChan {
				got = append(got, val)
				closer(val)
			}

			if !reflect.DeepEqual(got, ttt.Want) {
				t.Errorf("%s() = %v, want %v", name, got, ttt.Want)
			}
		})
	}
}
