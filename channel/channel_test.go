package channel

import (
	"testing"

	ezt "github.com/ezotaka/golib/testing"
)

func TestToChan(t *testing.T) {
	type args struct {
		done   <-chan any
		values []int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "length=0",
			Args: args{
				done:   make(<-chan any),
				values: []int{},
			},
			Want: []int{},
		},
		{
			Name: "length=1",
			Args: args{
				done:   make(<-chan any),
				values: []int{1},
			},
			Want: []int{1},
		},
		{
			Name: "length=2",
			Args: args{
				done:   make(<-chan any),
				values: []int{1, 2},
			},
			Want: []int{1, 2},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "ToChan", ToChan(a.done, a.values...)
	}

	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestRepeat(t *testing.T) {
	type args struct {
		done   <-chan any
		values []int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "1",
			Args: args{
				done:   make(<-chan any),
				values: []int{1},
			},
			Want: []int{1, 1, 1, 1, 1},
		},
		{
			Name: "1,2",
			Args: args{
				done:   make(<-chan any),
				values: []int{1, 2},
			},
			Want: []int{1, 2, 1, 2, 1},
		},
		{
			Name: "1,2,3,4,5,6",
			Args: args{
				done:   make(<-chan any),
				values: []int{1, 2, 3, 4, 5, 6},
			},
			Want: []int{1, 2, 3, 4, 5},
		},
		{
			Name: "no params",
			Args: args{
				done:   make(<-chan any),
				values: []int{},
			},
			Want: []int{0, 0, 0, 0, 0},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "Repeat", Take(a.done, Repeat(a.done, a.values...), 5)
	}
	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestRepeatFunc(t *testing.T) {
	counter := 0
	increment := func() int {
		counter++
		return counter
	}
	type args struct {
		done <-chan any
		fn   func() int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "5 count",
			Args: args{
				done: make(<-chan any),
				fn:   increment,
			},
			Want: []int{1, 2, 3, 4, 5},
		},
		{
			Name: "nil func",
			Args: args{
				done: make(<-chan any),
				fn:   nil,
			},
			Want: []int{0, 0, 0, 0, 0},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "RepeatFunc", Take(a.done, RepeatFunc(a.done, a.fn), 5)
	}
	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestTake(t *testing.T) {
	type args struct {
		done      <-chan any
		valueChan <-chan int
		num       int
	}
	done := make(chan any)
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "take 1",
			Args: args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       1,
			},
			Want: []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       2,
			},
			Want: []int{1, 2},
		},
		{
			Name: "try to take closed",
			Args: args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       3,
			},
			Want: []int{1, 2, 0},
		},
		{
			Name: "nil channel",
			Args: args{
				done:      done,
				valueChan: nil,
				num:       3,
			},
			Want: []int{},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "Take", Take(a.done, a.valueChan, a.num)
	}

	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}
