package chantest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestContextWithCountCancel(t *testing.T) {
	type args struct {
		cnt int
	}
	tests := []struct {
		name     string
		args     args
		want     int
		isPanic  bool
		panicMsg string
	}{
		{
			name: "cnt = 0",
			args: args{
				cnt: 0,
			},
			want: 0,
		},
		{
			name: "cnt = 1",
			args: args{
				cnt: 1,
			},
			want: 1,
		},
		{
			name: "cnt = -1",
			args: args{
				cnt: -1,
			},
			isPanic:  true,
			panicMsg: "cnt must be zero or positive",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); tt.isPanic && r == nil {
					t.Errorf("ContextWithCountCancel() doesn't panic, want '%v'", tt.panicMsg)
				} else if tt.isPanic && r != tt.panicMsg {
					t.Errorf("ContextWithCountCancel() panic '%v', want '%v'", r, tt.panicMsg)
				}
			}()
			got := ContextWithCountCancel(tt.args.cnt)
			if cnt, ok := countToCancel(got); !ok || !reflect.DeepEqual(cnt, tt.want) {
				t.Errorf("ContextWithCountCancel() got = %v, want %v", cnt, tt.want)
			}
		})
	}
}

const (
	// If TestSleep fails, increase this value.
	// The test will stabilize, but will take longer.
	timeScale = 20
)

func scaledTime(t float64) time.Duration {
	return time.Duration(t*timeScale) * time.Millisecond
}

func TestRun(t *testing.T) {
	// count up every 100 msec until end
	counter := func(end int) (<-chan int, error) {
		if end < 0 {
			return nil, errors.New("end must be zero or positive")
		}
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				time.Sleep(scaledTime(100))
				valChan <- count
				if count >= end {
					return
				}
				count++
			}
		}()
		return valChan, nil
	}
	counterCaller := func(_ context.Context, end int) (<-chan int, error) {
		return counter(end)
	}
	type args struct {
		ctx    context.Context
		fn     func(context.Context, int) (<-chan int, error)
		fnArgs int
	}
	tests := []struct { // TODO: use Case[C, A]
		name       string
		args       args
		want       []int
		errMsg     string
		panicValue any
	}{
		{
			name: "channel closed normally",
			args: args{
				ctx:    context.Background(),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 0",
			args: args{
				ctx:    ContextWithCountCancel(0),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{},
		},
		{
			name: "channel closed by context with count 1",
			args: args{
				ctx:    ContextWithCountCancel(1),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1},
		},
		{
			name: "channel closed by context with count 2",
			args: args{
				ctx:    ContextWithCountCancel(2),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "channel closed by context with count 3",
			args: args{
				ctx:    ContextWithCountCancel(3),
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		// TODO: timeout test is instability
		// {
		// 	name: "channel closed by timeout 0",
		// 	args: args{
		// 		ctx:    ContextWithTimeout(scaledTime(50)),
		// 		fn:     counterCaller,
		// 		fnArgs: 2,
		// 	},
		// 	want: []int{},
		// },
		// {
		// 	name: "channel closed by timeout 1",
		// 	args: args{
		// 		ctx:    ContextWithTimeout(scaledTime(150)),
		// 		fn:     counterCaller,
		// 		fnArgs: 2,
		// 	},
		// 	want: []int{1},
		// },
		// {
		// 	name: "channel closed by timeout 2",
		// 	args: args{
		// 		ctx:    ContextWithTimeout(scaledTime(250)),
		// 		fn:     counterCaller,
		// 		fnArgs: 2,
		// 	},
		// 	want: []int{1, 2},
		// },
		{
			name: "nil context is treated as context.Background()",
			args: args{
				ctx:    nil,
				fn:     counterCaller,
				fnArgs: 2,
			},
			want: []int{1, 2},
		},
		{
			name: "catch error caused by the function to be tested",
			args: args{
				ctx:    context.Background(),
				fn:     counterCaller,
				fnArgs: -1,
			},
			errMsg: "end must be zero or positive",
		},
		{
			name: "nil fn causes panic",
			args: args{
				ctx: context.Background(),
				fn:  nil,
			},
			panicValue: "c.Invoker must not be nil",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got []int
			var panicVal any
			var errMsg string

			defer func() {
				// process panic caused directory by Run function
				if r := recover(); r != nil {
					panicVal = r
				}

				if tt.panicValue != nil {
					if panicVal == nil {
						t.Errorf("Run() doesn't panic, want '%v'", tt.panicValue)
					} else if panicVal != tt.panicValue {
						t.Errorf("Run() panic '%v', want '%v'", panicVal, tt.panicValue)
					}
				} else if tt.errMsg != "" {
					if errMsg == "" {
						t.Errorf("Run() doesn't return error, want '%v'", tt.errMsg)
					} else if errMsg != tt.errMsg {
						t.Errorf("Run() returns '%v', want '%v'", errMsg, tt.errMsg)
					}
				} else if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Run() = %v, want %v", got, tt.want)
				}
			}()

			c := Case[int, int]{
				Name:    tt.name,
				Args:    tt.args.fnArgs,
				Context: tt.args.ctx,
				Invoker: tt.args.fn,
				Want:    tt.want,
			}
			got, errMsg, panicVal = Run(c)
		})
	}
}
