package channel

import (
	"testing"

	ezt "github.com/ezotaka/golib/testing"
)

func TestToChan(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "length=0",
			Args: args{
				values: []int{},
			},
			Want: []int{},
		},
		{
			Name: "length=1",
			Args: args{
				values: []int{1},
			},
			Want: []int{1},
		},
		{
			Name: "length=2",
			Args: args{
				values: []int{1, 2},
			},
			Want: []int{1, 2},
		},
	}
	call := func(done <-chan any, a args) (string, <-chan int) {
		return "ToChan", ToChan(done, a.values...)
	}

	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestRepeat(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "1",
			Args: args{
				values: []int{1},
			},
			Want: []int{1, 1, 1, 1, 1},
		},
		{
			Name: "1,2",
			Args: args{
				values: []int{1, 2},
			},
			Want: []int{1, 2, 1, 2, 1},
		},
		{
			Name: "1,2,3,4,5,6",
			Args: args{
				values: []int{1, 2, 3, 4, 5, 6},
			},
			Want: []int{1, 2, 3, 4, 5},
		},
		{
			Name: "no params",
			Args: args{
				values: []int{},
			},
			Want: []int{0, 0, 0, 0, 0},
		},
	}
	call := func(done <-chan any, a args) (string, <-chan int) {
		return "Repeat", Take(done, Repeat(done, a.values...), 5)
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
		fn func() int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "5 count",
			Args: args{
				fn: increment,
			},
			Want: []int{1, 2, 3, 4, 5},
		},
		{
			Name: "nil func",
			Args: args{
				fn: nil,
			},
			Want: []int{0, 0, 0, 0, 0},
		},
	}
	call := func(done <-chan any, a args) (string, <-chan int) {
		return "RepeatFunc", Take(done, RepeatFunc(done, a.fn), 5)
	}
	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}

func TestTake(t *testing.T) {
	type args struct {
		valueChan <-chan int
		num       int
	}
	tests := []ezt.ChanFuncTestCase[int, args]{
		{
			Name: "take 1",
			Args: args{
				valueChan: ToChan(make(<-chan any), 1, 2),
				num:       1,
			},
			Want: []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				valueChan: ToChan(make(<-chan any), 1, 2),
				num:       2,
			},
			Want: []int{1, 2},
		},
		{
			Name: "try to take closed",
			Args: args{
				valueChan: ToChan(make(<-chan any), 1, 2),
				num:       3,
			},
			Want: []int{1, 2, 0},
		},
		{
			Name: "nil channel",
			Args: args{
				valueChan: nil,
				num:       3,
			},
			Want: []int{},
		},
	}
	call := func(done <-chan any, a args) (string, <-chan int) {
		return "Take", Take(done, a.valueChan, a.num)
	}

	ezt.ExecReadOnlyChanFuncTest(t, tests, call)
}
