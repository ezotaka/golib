package channel

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ezotaka/golib/conv"
	"github.com/ezotaka/golib/eztest"
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

func TestEnumerate(t *testing.T) {
	type args struct {
		c <-chan string
	}
	closed := make(chan string)
	close(closed)
	tests := []struct {
		name string
		args args
		want []forEnum[string]
	}{
		{
			name: "normal",
			args: args{
				c: conv.Chan("a", "b"),
			},
			want: []forEnum[string]{
				{
					I: 0,
					V: "a",
				},
				{
					I: 1,
					V: "b",
				},
			},
		},
		{
			name: "closed channel",
			args: args{
				c: closed,
			},
			want: []forEnum[string]{},
		},
		{
			name: "nil channel",
			args: args{
				c: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conv.Slice(Enumerate(tt.args.c)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrDone(t *testing.T) {
	type args struct {
		channel <-chan int
	}
	invoker := eztest.Invoker[args, <-chan int]{
		Name: "OrDone",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return OrDone(ctx, a.channel), nil
		},
	}
	tests := []eztest.Case[args, <-chan int, []int]{
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
			Context: eztest.ContextWithCountCancel(2),
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "blocked by reading nil channel, but done after 100 msec",
			Args: args{
				channel: nil,
			},
			Context: eztest.ContextWithTimeout(100 * time.Millisecond),
			Invoker: invoker,
			Want:    []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := RunTest(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	type args struct {
		values []int
	}
	invoker := eztest.Invoker[args, <-chan int]{
		Name: "Repeat",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Repeat(ctx, a.values...), nil
		},
	}
	tests := []eztest.Case[args, <-chan int, []int]{
		{
			Name: "1",
			Args: args{
				values: []int{1},
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 1, 1, 1, 1},
		},
		{
			Name: "1,2",
			Args: args{
				values: []int{1, 2},
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 1, 2, 1},
		},
		{
			Name: "1,2,3,4,5,6",
			Args: args{
				values: []int{1, 2, 3, 4, 5, 6},
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 3, 4, 5},
		},
		{
			Name: "no params",
			Args: args{
				values: []int{},
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := RunTest(tt); err != nil {
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
	invoker := eztest.Invoker[args, <-chan int]{
		Name: "RepeatFunc",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return RepeatFunc(ctx, a.fn), nil
		},
	}
	tests := []eztest.Case[args, <-chan int, []int]{
		{
			Name: "5 count",
			Args: args{
				fn: increment,
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Want:    []int{1, 2, 3, 4, 5},
		},
		{
			Name: "nil func",
			Args: args{
				fn: nil,
			},
			Context: eztest.ContextWithCountCancel(5),
			Invoker: invoker,
			Panic:   "fn must not be nil",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := RunTest(tt); err != nil {
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
	invoker := eztest.Invoker[args, <-chan int]{
		Name: "Take",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Take(ctx, a.valueChan, a.num), nil
		},
	}

	tests := []eztest.Case[args, <-chan int, []int]{
		{
			Name: "take 1",
			Args: args{
				valueChan: conv.Chan(1, 2),
				num:       1,
			},
			Invoker: invoker,
			Want:    []int{1},
		},
		{
			Name: "take 2",
			Args: args{
				valueChan: conv.Chan(1, 2),
				num:       2,
			},
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "can't take from closed channel",
			Args: args{
				valueChan: conv.Chan(1, 2),
				num:       3,
			},
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "try to take 3 but canceled at 2",
			Args: args{
				valueChan: conv.Chan(1, 2, 3, 4),
				num:       3,
			},
			Context: eztest.ContextWithCountCancel(2),
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
			Want:    nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := RunTest(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

func TestSleep(t *testing.T) {
	type args struct {
		c <-chan int
		t time.Duration
	}
	invoker := eztest.Invoker[args, <-chan int]{
		Name: "Sleep",
		Invoke: func(ctx context.Context, a args) (<-chan int, error) {
			return Sleep(ctx, a.c, a.t), nil
		},
	}
	tests := []eztest.Case[args, <-chan int, []int]{
		{
			Name: "no count",
			Args: args{
				c: conv.Chan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: eztest.ContextWithTimeout(scaledTime(5)),
			Invoker: invoker,
			Want:    []int{},
		},
		{
			Name: "sleep 1",
			Args: args{
				c: conv.Chan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: eztest.ContextWithTimeout(scaledTime(150)),
			Invoker: invoker,
			Want:    []int{1},
		},
		{
			Name: "sleep 2",
			Args: args{
				c: conv.Chan(1, 2, 3),
				t: scaledTime(100),
			},
			Context: eztest.ContextWithTimeout(scaledTime(250)),
			Invoker: invoker,
			Want:    []int{1, 2},
		},
		{
			Name: "0 time sleep",
			Args: args{
				c: conv.Chan(1, 2, 3),
				t: scaledTime(0),
			},
			// nervousness
			Context: eztest.ContextWithTimeout(scaledTime(10)),
			Invoker: invoker,
			Want:    []int{1, 2, 3},
		},
		{
			Name: "cannot read (total sleep > deadline)",
			Args: args{
				c: Sleep(
					context.Background(),
					conv.Chan(1, 2, 3),
					scaledTime(120),
				),
				t: scaledTime(100),
			},
			//Non-determined
			Context: eztest.ContextWithTimeout(scaledTime(210)),
			Invoker: invoker,
			Want:    []int{},
		},
		{
			Name: "can read (total sleep < deadline)",
			Args: args{
				c: Sleep(
					context.Background(),
					conv.Chan(1, 2, 3),
					scaledTime(120),
				),
				t: scaledTime(100),
			},
			//Non-determined
			Context: eztest.ContextWithTimeout(scaledTime(240)),
			Invoker: invoker,
			Want:    []int{1},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := RunTest(tt); err != nil {
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
// 	tests := []eztest.Case[int, args]{
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
