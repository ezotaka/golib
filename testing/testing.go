package testing

import (
	"reflect"
	"testing"
)

type ChanFuncTestCase[C any, A any] struct {
	Name string
	Args A
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
