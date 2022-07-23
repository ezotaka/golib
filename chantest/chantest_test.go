package chantest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ezotaka/golib/eztest"
)

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
	counterInvoker := eztest.Invoker[counterArgs, <-chan int]{
		Name: "counter",
		Invoke: func(ctx context.Context, a counterArgs) (<-chan int, error) {
			return counter(ctx, a.end)
		},
	}

	// used for test of RunChannel
	type runChannelArgs struct {
		tc eztest.Case[counterArgs, <-chan int, []int]
	}
	runChannelInvoker := eztest.Invoker[runChannelArgs, []int]{
		Name: "RunChannel",
		Invoke: func(_ context.Context, a runChannelArgs) ([]int, error) {
			return Run(a.tc)
		},
	}

	// shortcut value
	ints1 := []int{1}
	ints2 := []int{1, 2}

	tests := []eztest.Case[runChannelArgs, []int, []int]{
		{
			Name: "OK closed normally",
			Args: runChannelArgs{
				tc: eztest.Case[counterArgs, <-chan int, []int]{
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
			Name: "OK closed by context",
			Args: runChannelArgs{
				tc: eztest.Case[counterArgs, <-chan int, []int]{
					Name:    "channel closed by context",
					Context: eztest.ContextWithCountCancel(1),
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
			Name: "OK error",
			Args: runChannelArgs{
				tc: eztest.Case[counterArgs, <-chan int, []int]{
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
			Name: "OK panic",
			Args: runChannelArgs{
				tc: eztest.Case[counterArgs, <-chan int, []int]{
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
	}

	// simple post processor that just returns the arguments as they are
	pp := func(ctx context.Context, r []int, err error) ([]int, error) {
		return r, err
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			if _, err := eztest.Run(tt, pp); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
