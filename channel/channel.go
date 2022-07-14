package channel

import (
	"reflect"
	"testing"
)

func ToChan[T any](done <-chan any, values ...T) <-chan T {
	ch := make(chan T, len(values))
	go func() {
		defer close(ch)
		for _, v := range values {
			select {
			case <-done:
				return
			case ch <- v:
			}
		}
	}()
	return ch
}

func Repeat[T any](done <-chan any, values ...T) <-chan T {
	valuesChan := make(chan T)
	go func() {
		defer close(valuesChan)
		if len(values) == 0 {
			return
		}
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valuesChan <- v:
				}
			}
		}
	}()
	return valuesChan
}

func RepeatFunc[T any](
	done <-chan any,
	fn func() T,
) <-chan T {
	valueChan := make(chan T)
	go func() {
		defer close(valueChan)
		if fn == nil {
			return
		}
		for {
			select {
			case <-done:
				return
			case valueChan <- fn():
			}
		}
	}()
	return valueChan
}

func Take[T any](
	done <-chan any,
	valueChan <-chan T,
	num int,
) <-chan T {
	takeChan := make(chan T)
	go func() {
		defer close(takeChan)
		if valueChan == nil {
			return
		}
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeChan <- <-valueChan:
			}
		}
	}()
	return takeChan
}

type ChanFuncTestCase[C any, A any] struct {
	Name string
	Args A
	Want []C
}

// Execute Table Driven Test for function which returns <-chan TC
func ExecReadOnlyChanFuncTest[
	C any,
	A any,
	T ChanFuncTestCase[C, A],
](
	t *testing.T,
	tests []T,
	call func(A) (string, <-chan C),
) {
	for _, tt := range tests {
		ttt := ChanFuncTestCase[C, A](tt)
		t.Run(ttt.Name, func(t *testing.T) {
			got := []C{}
			name, gotChan := call(ttt.Args)
			for g := range gotChan {
				got = append(got, g)
			}
			if !reflect.DeepEqual(got, ttt.Want) {
				t.Errorf("%s() = %v, want %v", name, got, ttt.Want)
			}
		})
	}
}
