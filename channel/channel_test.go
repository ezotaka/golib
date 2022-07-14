package channel

import (
	"testing"
)

func TestToChan(t *testing.T) {
	type args struct {
		done   <-chan any
		values []int
	}
	tests := []ChanFuncTestCase[int, args]{
		{
			"length=0",
			args{
				done:   make(<-chan any),
				values: []int{},
			},
			[]int{},
		},
		{
			"length=1",
			args{
				done:   make(<-chan any),
				values: []int{1},
			},
			[]int{1},
		},
		{
			"length=2",
			args{
				done:   make(<-chan any),
				values: []int{1, 2},
			},
			[]int{1, 2},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "ToChan", ToChan(a.done, a.values...)
	}

	ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestRepeat(t *testing.T) {
	type args struct {
		done   <-chan any
		values []int
	}
	tests := []ChanFuncTestCase[int, args]{
		{
			"1",
			args{
				done:   make(<-chan any),
				values: []int{1},
			},
			[]int{1, 1, 1, 1, 1},
		},
		{
			"1,2",
			args{
				done:   make(<-chan any),
				values: []int{1, 2},
			},
			[]int{1, 2, 1, 2, 1},
		},
		{
			"1,2,3,4,5,6",
			args{
				done:   make(<-chan any),
				values: []int{1, 2, 3, 4, 5, 6},
			},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"no params",
			args{
				done:   make(<-chan any),
				values: []int{},
			},
			[]int{0, 0, 0, 0, 0},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "Repeat", Take(a.done, Repeat(a.done, a.values...), 5)
	}
	ExecReadOnlyChanFuncTest(t, tests, call)
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
	tests := []ChanFuncTestCase[int, args]{
		{
			"5 count",
			args{
				done: make(<-chan any),
				fn:   increment,
			},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"nil func",
			args{
				done: make(<-chan any),
				fn:   nil,
			},
			[]int{0, 0, 0, 0, 0},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "RepeatFunc", Take(a.done, RepeatFunc(a.done, a.fn), 5)
	}
	ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestTake(t *testing.T) {
	type args struct {
		done      <-chan any
		valueChan <-chan int
		num       int
	}
	done := make(chan any)
	tests := []ChanFuncTestCase[int, args]{
		{
			"take 1",
			args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       1,
			},
			[]int{1},
		},
		{
			"take 2",
			args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       2,
			},
			[]int{1, 2},
		},
		{
			"try to take closed",
			args{
				done:      done,
				valueChan: ToChan(done, 1, 2),
				num:       3,
			},
			[]int{1, 2, 0},
		},
		{
			"nil channel",
			args{
				done:      done,
				valueChan: nil,
				num:       3,
			},
			[]int{},
		},
	}
	call := func(a args) (string, <-chan int) {
		return "Take", Take(a.done, a.valueChan, a.num)
	}

	ExecReadOnlyChanFuncTest(t, tests, call)
}
