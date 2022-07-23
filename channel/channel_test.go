package channel

import (
	"context"
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
	invoker := chantest.Invoker[args, <-chan int]{
		Name: "OrDone",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return OrDone(ctx, a.channel), nil
		},
	}
	tests := []chantest.Case[args, <-chan int, []int]{
		{
			Name: "no interruption due to done",
			Args: args{
				channel: counter(3),
			},
			Invoker: invoker,
			Want:    []int{1, 2, 3},
		},
		{
			Name: "interruption due to done",
			Args: args{
				channel: counter(3),
			},
			Context: chantest.ContextWithCountCancel(2),
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "blocked by reading nil channel, but done after 100 msec",
			Args: args{
				channel: nil,
			},
			Context: chantest.ContextWithTimeout(100 * time.Millisecond),
			Invoker: invoker,
			Want:    []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := chantest.RunChannel(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	type args struct {
		values []int
	}
	invoker := chantest.Invoker[args, <-chan int]{
		Name: "Repeat",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Repeat(ctx, a.values...), nil
		},
	}
	tests := []chantest.Case[args, <-chan int, []int]{
		{
			Name: "1",
			Args: args{
				values: []int{1},
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 1, 1, 1, 1},
		},
		{
			Name: "1,2",
			Args: args{
				values: []int{1, 2},
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 1, 2, 1},
		},
		{
			Name: "1,2,3,4,5,6",
			Args: args{
				values: []int{1, 2, 3, 4, 5, 6},
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 3, 4, 5},
		},
		{
			Name: "no params",
			Args: args{
				values: []int{},
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := chantest.RunChannel(tt); err != nil {
				t.Errorf(err.Error())
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
	invoker := chantest.Invoker[args, <-chan int]{
		Name: "RepeatFunc",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return RepeatFunc(ctx, a.fn), nil
		},
	}
	tests := []chantest.Case[args, <-chan int, []int]{
		{
			Name: "5 count",
			Args: args{
				fn: increment,
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 3, 4, 5},
		},
		{
			Name: "nil func",
			Args: args{
				fn: nil,
			},
			Context: chantest.ContextWithCountCancel(5),
			Invoker: invoker,
			Panic:   "fn must not be nil",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := chantest.RunChannel(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestTake(t *testing.T) {
	type args struct {
		valueChan <-chan int
		num       int
	}
	invoker := chantest.Invoker[args, <-chan int]{
		Name: "Take",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Take(ctx, a.valueChan, a.num), nil
		},
	}

	tests := []chantest.Case[args, <-chan int, []int]{
		{
			Name: "take 1",
			Args: args{
				valueChan: conv.ToChan(1, 2),
				num:       1,
			},
			Invoker: invoker,
			Want:    []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				valueChan: conv.ToChan(1, 2),
				num:       2,
			},
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "try to take closed",
			Args: args{
				valueChan: conv.ToChan(1, 2),
				num:       3,
			},
			Invoker: invoker,
			Want:    []int{1, 2, 0},
		},
		{
			Name: "try to take 3 but canceled at 2",
			Args: args{
				valueChan: conv.ToChan(1, 2, 3, 4),
				num:       3,
			},
			Context: chantest.ContextWithCountCancel(2),
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "nil channel",
			Args: args{
				valueChan: nil,
				num:       3,
			},
			Invoker: invoker,
			Want:    []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := chantest.RunChannel(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

const (
	// If TestSleep fails, increase this value.
	// The test will stabilize, but will take longer.
	timeScale = 5
)

func scaledTime(t float64) time.Duration {
	return time.Duration(t*timeScale) * time.Millisecond
}

func TestSleep(t *testing.T) {
	type args struct {
		c <-chan int
		t time.Duration
	}
	invoker := chantest.Invoker[args, <-chan int]{
		Name: "Sleep",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Sleep(ctx, a.c, a.t), nil
		},
	}
	tests := []chantest.Case[args, <-chan int, []int]{
		{
			Name: "no count",
			Args: args{
				c: conv.ToChan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: chantest.ContextWithTimeout(scaledTime(5)),
			Invoker: invoker,
			Want:    []int{},
		},
		{
			Name: "sleep 1",
			Args: args{
				c: conv.ToChan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: chantest.ContextWithTimeout(scaledTime(150)),
			Invoker: invoker,
			Want:    []int{1},
		},
		{
			Name: "sleep 2",
			Args: args{
				c: conv.ToChan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: chantest.ContextWithTimeout(scaledTime(250)),
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "0 time sleep",
			Args: args{
				c: conv.ToChan(1, 2, 3),
				t: scaledTime(0),
			},
			// nervousness
			Context: chantest.ContextWithTimeout(scaledTime(10)),
			Invoker: invoker,
			Want:    []int{1, 2, 3},
		},
		{
			Name: "cannot read (total sleep > deadline)",
			Args: args{
				c: Sleep(
					context.Background(),
					conv.ToChan(1, 2, 3),
					scaledTime(120),
				),
				t: scaledTime(100),
			},
			//Non-determined
			Context: chantest.ContextWithTimeout(scaledTime(210)),
			Invoker: invoker,
			Want:    []int{},
		},
		{
			Name: "can read (total sleep < deadline)",
			Args: args{
				c: Sleep(
					context.Background(),
					conv.ToChan(1, 2, 3),
					scaledTime(120),
				),
				t: scaledTime(100),
			},
			//Non-determined
			Context: chantest.ContextWithTimeout(scaledTime(240)),
			Invoker: invoker,
			Want:    []int{1},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := chantest.RunChannel(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

// TODO: Write test
// func TestTee(t *testing.T) {
// 	type args struct {
// 		in  <-chan int
// 		num int
// 	}
// 	tests := []chantest.Case[int, args]{
// 		{
// 			Name: "length = 2",
// 			Args: args{
// 				in:  conv.ToChan(1, 2),
// 				num: 1,
// 			},
// 			Want: []int{1, 2},
// 		},
// 		// TODO: test below doesn't work
// 		// {
// 		// 	Name: "canceled",
// 		// 	Args: args{
// 		// 		in:  ToChan(1, 2),
// 		// 		num: 1,
// 		// 	},
// 		// 	IsDoneByIndex: func(i int) bool { return i == 0 },
// 		// 	Want:          []int{1},
// 		// },
// 	}
// 	caller := func(ctx context.Context, a args) (<-chan int, <-chan int) {
// 		return Tee(ctx, a.in)
// 	}
// 	for _, tt := range tests {
// 		tt := tt
// 		t.Run(tt.Name, func(t *testing.T) {
// 			t.Parallel()
// 			got1, got2 := chantest.GetTestedValues2(tt, caller)
// 			if !reflect.DeepEqual(got1, tt.Want) {
// 				t.Errorf("Take() got1 = %v, want %v", got1, tt.Want)
// 			}
// 			if !reflect.DeepEqual(got2, tt.Want) {
// 				t.Errorf("Take() got2 = %v, want %v", got2, tt.Want)
// 			}
// 		})
// 	}
// }
