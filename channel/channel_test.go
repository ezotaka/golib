package channel

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ezotaka/golib/chantest"
	"github.com/ezotaka/golib/conv"
)

// function to be tested
// count up every 100 msec until end
func counter(end int) <-chan int {
	valChan := make(chan int)
	go func() {
		defer close(valChan)
		count := 1
		for {
			time.Sleep(100 * time.Millisecond)
			valChan <- count
			if count >= end {
				return
			}
			count++
		}
	}()
	return valChan
}

func TestOrDone(t *testing.T) {
	type args struct {
		channel <-chan int
	}
	tests := []chantest.Case[int, args]{
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
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("OrDone() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	type args struct {
		values []int
	}
	tests := []chantest.Case[int, args]{
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
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
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
	tests := []chantest.Case[int, args]{
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
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
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
	tests := []chantest.Case[int, args]{
		{
			Name: "take 1",
			Args: args{
				valueChan: conv.ToChan(1, 2),
				num:       1,
			},
			Want: []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				valueChan: conv.ToChan(1, 2),
				num:       2,
			},
			Want: []int{1, 2},
		},
		{
			Name: "try to take closed",
			Args: args{
				valueChan: conv.ToChan(1, 2),
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
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("Take() = %v, want %v", got, tt.Want)
			}
		})
	}
}

const (
	// If TestSleep fails, increase this value.
	// The test will stabilize, but will take longer.
	sleepTestScale = 5
)

func sleepTime(t float64) time.Duration {
	return time.Duration(t*sleepTestScale) * time.Millisecond
}

func TestSleep(t *testing.T) {
	type args struct {
		c func(context.Context) <-chan int
		t time.Duration
	}
	tests := []chantest.Case[int, args]{
		{
			Name: "no count",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return conv.ToChan(1, 2, 3)
				},
				t: sleepTime(100),
			},
			IsDoneByTime: sleepTime(5),
			Want:         []int{},
		},
		{
			Name: "sleep 1",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return conv.ToChan(1, 2, 3)
				},
				t: sleepTime(100),
			},
			IsDoneByTime: sleepTime(150),
			Want:         []int{1},
		},
		{
			Name: "sleep 2",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return conv.ToChan(1, 2, 3)
				},
				t: sleepTime(100),
			},
			IsDoneByTime: sleepTime(250),
			Want:         []int{1, 2},
		},
		{
			Name: "0 time sleep",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return conv.ToChan(1, 2, 3)
				},
				t: sleepTime(0),
			},
			// nervousness
			IsDoneByTime: sleepTime(10),
			Want:         []int{1, 2, 3},
		},
		{
			Name: "cannot read (total sleep > deadline)",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return Sleep(
						ctx,
						conv.ToChan(1, 2, 3),
						sleepTime(120))
				},
				t: sleepTime(100),
			},
			//Non-determined
			IsDoneByTime: sleepTime(210),
			Want:         []int{},
		},
		{
			Name: "can read (total sleep < deadline)",
			Args: args{
				c: func(ctx context.Context) <-chan int {
					return Sleep(
						ctx,
						conv.ToChan(1, 2, 3),
						sleepTime(120))
				},
				t: sleepTime(100),
			},
			//Non-determined
			IsDoneByTime: sleepTime(240),
			Want:         []int{1},
		},
	}
	caller := func(ctx context.Context, a args) <-chan int {
		return Sleep(ctx, (a.c)(ctx), a.t)
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got := chantest.GetTestedValues(tt, caller)
			if !reflect.DeepEqual(got, tt.Want) {
				t.Errorf("Sleep() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestTee(t *testing.T) {
	type args struct {
		in  <-chan int
		num int
	}
	tests := []chantest.Case[int, args]{
		{
			Name: "length = 2",
			Args: args{
				in:  conv.ToChan(1, 2),
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
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := chantest.GetTestedValues2(tt, caller)
			if !reflect.DeepEqual(got1, tt.Want) {
				t.Errorf("Take() got1 = %v, want %v", got1, tt.Want)
			}
			if !reflect.DeepEqual(got2, tt.Want) {
				t.Errorf("Take() got2 = %v, want %v", got2, tt.Want)
			}
		})
	}
}
