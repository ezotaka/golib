package channel

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestOrDone(t *testing.T) {
	// function to be tested
	// count up every 100 msec until end
	counter := func(end int) <-chan int {
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				time.Sleep(100 * time.Microsecond)
				valChan <- count
				if count >= end {
					return
				}
				count++
			}
		}()
		return valChan
	}

	type args struct {
		channel <-chan int
	}
	tests := []TestCase[int, args]{
		{
			Name: "no interruption due to done",
			Args: args{
				channel: counter(3),
			},
			Want: []int{1, 2, 3},
		},
		{
			Name:          "interruption due to done",
			IsDoneByValue: func(v int) bool { return v == 2 },
			Args: args{
				channel: counter(3),
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
	caller := func(ctx context.Context, a args) <-chan int {
		return OrDone(ctx, a.channel)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := GetTestedValues(tt, caller)
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
	caller := func(ctx context.Context, a args) <-chan int {
		return ToChan(a.values...)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := GetTestedValues(tt, caller)
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
	caller := func(ctx context.Context, a args) <-chan int {
		return Take(ctx, Repeat(ctx, a.values...), 5)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := GetTestedValues(tt, caller)
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
	caller := func(ctx context.Context, a args) <-chan int {
		return Take(ctx, RepeatFunc(ctx, a.fn), 5)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := GetTestedValues(tt, caller)
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
				valueChan: ToChan(1, 2),
				num:       1,
			},
			Want: []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				valueChan: ToChan(1, 2),
				num:       2,
			},
			Want: []int{1, 2},
		},
		{
			Name: "try to take closed",
			Args: args{
				valueChan: ToChan(1, 2),
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
	caller := func(ctx context.Context, a args) <-chan int {
		return Take(ctx, a.valueChan, a.num)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := GetTestedValues(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("Take() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestTee(t *testing.T) {
	type args struct {
		in  <-chan int
		num int
	}
	tests := []TestCase[int, args]{
		{
			Name: "length = 2",
			Args: args{
				in:  ToChan(1, 2),
				num: 1,
			},
			Want: []int{1, 2},
		},
		// TODO: test below doesn't work
		// {
		// 	Name: "canceled",
		// 	Args: args{
		// 		in:  ToChan(1, 2),
		// 		num: 1,
		// 	},
		// 	IsDoneByIndex: func(i int) bool { return i == 0 },
		// 	Want:          []int{1},
		// },
	}
	caller := func(ctx context.Context, a args) (<-chan int, <-chan int) {
		return Tee(ctx, a.in)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := GetTestedValues2(tt, caller)
			if !reflect.DeepEqual(got1, tt.Want) {
				t.Errorf("Take() got1 = %v, want %v", got1, tt.Want)
			}
			if !reflect.DeepEqual(got2, tt.Want) {
				t.Errorf("Take() got2 = %v, want %v", got2, tt.Want)
			}
		})
	}
}
