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

func TestRunChannel(t *testing.T) {
	const errMsgOK = "end is zero"
	const panicOK = "end is negative"
	const errMsgNG = "NG ErrMsg"
	const panicNG = "NG Panic"

	// counter function is tested by RunChannel function
	//
	// return channel which sends [1, 2, ..., end] every scaledTime(100)
	// error if end == 0
	// panic if end < 0
	counter := func(ctx context.Context, end int) (<-chan int, error) {
		if end == 0 {
			return nil, errors.New(errMsgOK)
		} else if end < 0 {
			panic(panicOK)
		}
		valChan := make(chan int)
		go func() {
			defer close(valChan)
			count := 1
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(scaledTime(100)):
					valChan <- count
					if count >= end {
						return
					}
					count++
				}
			}
		}()
		return valChan, nil
	}
	type counterArgs struct {
		end int
	}
	counterInvoker := Invoker[counterArgs, <-chan int]{
		Name: "counter",
		Invoke: func(ctx context.Context, a counterArgs) (<-chan int, error) {
			return counter(ctx, a.end)
		},
	}

	// used for test of RunChannel
	type runChannelArgs struct {
		tc Case[counterArgs, <-chan int, []int]
	}
	runChannelInvoker := Invoker[runChannelArgs, []int]{
		Name: "RunChannel",
		Invoke: func(_ context.Context, a runChannelArgs) ([]int, error) {
			return RunChannel(a.tc)
		},
	}

	// shortcut value
	ints1 := []int{1}
	ints2 := []int{1, 2}

	tests := []Case[runChannelArgs, []int, []int]{
		{
			Name: "OK closed normally",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "channel closed normally",
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints2,
				},
			},
			Invoker: runChannelInvoker,
			Want:    ints2,
		},
		{
			Name: "NG wrong Want, closed normally",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "channel closed normally",
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints1,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  notEqualsMsg(counterInvoker.Name, ints2, ints1),
		},
		{
			Name: "OK closed by context",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name:    "channel closed by context",
					Context: ContextWithCountCancel(1),
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints1,
				},
			},
			Invoker: runChannelInvoker,
			Want:    ints1,
		},
		{
			Name: "NG wrong Want, closed by context",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name:    "channel closed by context",
					Context: ContextWithCountCancel(1),
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints2,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  notEqualsMsg(counterInvoker.Name, ints1, ints2),
		},
		{
			Name: "OK error",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = 0",
					Args: counterArgs{
						end: 0,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgOK,
				},
			},
			Invoker: runChannelInvoker,
		},
		{
			Name: "NG wrong error",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = 0",
					Args: counterArgs{
						end: 0,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgNG,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  wrongErrorMsg(counterInvoker.Name, errMsgOK, errMsgNG),
		},
		{
			Name: "NG not error",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = 1",
					Args: counterArgs{
						end: 1,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgOK,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  notErrorMsg(counterInvoker.Name, errMsgOK),
		},

		{
			Name: "OK panic",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = -1",
					Args: counterArgs{
						end: -1,
					},
					Invoker: counterInvoker,
					Panic:   panicOK,
				},
			},
			Invoker: runChannelInvoker,
		},
		{
			Name: "NG wrong panic",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = -1",
					Args: counterArgs{
						end: -1,
					},
					Invoker: counterInvoker,
					Panic:   panicNG,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  wrongPanicMsg(counterInvoker.Name, panicOK, panicNG),
		},
		{
			Name: "NG not panic",
			Args: runChannelArgs{
				tc: Case[counterArgs, <-chan int, []int]{
					Name: "end = 1",
					Args: counterArgs{
						end: 1,
					},
					Invoker: counterInvoker,
					Panic:   panicOK,
				},
			},
			Invoker: runChannelInvoker,
			ErrMsg:  notPanicMsg(counterInvoker.Name, panicOK),
		},
		// TODO: not passed
		// {
		// 	Name: "PANIC invoke",
		// 	Args: runChannelArgs{
		// 		tc: Case[counterArgs, <-chan int, []int]{
		// 			Name: "end = 1",
		// 			Args: counterArgs{
		// 				end: 1,
		// 			},
		// 			Invoker: counterInvoker,
		// 			Panic:   panicOK,
		// 		},
		// 	},
		// 	Invoker: Invoker[runChannelArgs, []int]{
		// 		Name:   "Invoke is nil",
		// 		Invoke: nil,
		// 	},
		// 	Panic: "c.Invoker.Invoke must not be nil",
		// },
	}

	// simple post processor that just returns the arguments as they are
	pp := func(ctx context.Context, r []int, err error) ([]int, error) {
		return r, err
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := Run(tt, pp); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}

// TODO: write test
// func TestRun(t *testing.T) {
// 	type args struct {
// 		tc Case[A, R, W]
// 		pp func(context.Context, R, error) (W, error)
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantGot W
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotGot, err := Run(tt.args.tc, tt.args.pp)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(gotGot, tt.wantGot) {
// 				t.Errorf("Run() = %v, want %v", gotGot, tt.wantGot)
// 			}
// 		})
// 	}
// }
