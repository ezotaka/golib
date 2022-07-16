package channel

import (
	"reflect"
	"testing"
	"time"
)

func TestOrDone(t *testing.T) {
	type args struct {
		channel <-chan int
	}
	tests := []TestCase[int, args]{
		{
			Name: "no interruption due to done",
			Args: args{
				channel: ToChan(make(<-chan any), 1, 2, 3),
			},
			Want: []int{1, 2, 3},
		},
		{
			Name:          "interruption due to done",
			IsDoneByValue: func(v int) bool { return v == 2 },
			Args: args{
				channel: ToChan(make(<-chan any), 1, 2, 3),
			},
			Want: []int{1, 2},
		},
		{
			Name:         "blocked by reading nil channel, but done after 100 msec",
			IsDoneByTime: 100 * time.Millisecond,
			Args: args{
				channel: nil,
			},
			Want: []int{},
		},
	}
	caller := func(done <-chan any, a args) <-chan int {
		return OrDone(done, a.channel)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := DoTest(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("OrDone() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestToChan(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []TestCase[int, args]{
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
	caller := func(done <-chan any, a args) <-chan int {
		return ToChan(done, a.values...)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := DoTest(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("ToChan() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []TestCase[int, args]{
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
	caller := func(done <-chan any, a args) <-chan int {
		return Take(done, Repeat(done, a.values...), 5)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := DoTest(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("Repeat() = %v, want %v", got, tt.Want)
			}
		})
	}

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
	tests := []TestCase[int, args]{
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
	caller := func(done <-chan any, a args) <-chan int {
		return Take(done, RepeatFunc(done, a.fn), 5)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := DoTest(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("RepeatFunc() = %v, want %v", got, tt.Want)
			}
		})
	}

}

func TestTake(t *testing.T) {
	type args struct {
		valueChan <-chan int
		num       int
	}
	tests := []TestCase[int, args]{
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
	caller := func(done <-chan any, a args) <-chan int {
		return Take(done, a.valueChan, a.num)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := DoTest(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("Take() = %v, want %v", got, tt.Want)
			}
		})
	}
}
