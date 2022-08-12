// Copyright (c) 2022 Takatomo Ezo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eztest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
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
			if cnt, ok := CountToCancel(got); !ok || !reflect.DeepEqual(cnt, tt.want) {
				t.Errorf("ContextWithCountCancel() got = %v, want %v", cnt, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	const errMsgOK = "end is zero"
	const panicOK = "end is negative"
	const errMsgNG = "NG ErrMsg"
	const panicNG = "NG Panic"

	// counter function is tested by Run function
	//
	// return []int{1, 2, ... , end}
	// error if end == 0
	// panic if end < 0
	counter := func(ctx context.Context, end int) (ret []int, err error) {
		// ContextWithCountCancel can override end
		if cnt, ok := CountToCancel(ctx); ok {
			end = cnt
		}

		if end < 0 {
			panic(panicOK)
		}

		if end == 0 {
			err = errors.New(errMsgOK)
			return
		}

		for i := 1; i <= end; i++ {
			ret = append(ret, i)
		}
		return
	}
	type counterArgs struct {
		end int
	}
	counterInvoker := Invoker[counterArgs, []int]{
		Name: "counter",
		Invoke: func(ctx context.Context, a counterArgs) ([]int, error) {
			return counter(ctx, a.end)
		},
	}

	// used for test of Run
	type runArgs struct {
		tc Case[counterArgs, []int, string]
	}
	runInvoker := Invoker[runArgs, string]{
		Name: "Run",
		Invoke: func(_ context.Context, a runArgs) (string, error) {
			pp := func(_ context.Context, ints []int, err error) (string, error) {
				return fmt.Sprintf("%v", ints), err
			}
			return Run(a.tc, pp)
		},
	}

	// shortcut value
	ints1 := "[1]"
	ints2 := "[1 2]"

	tests := []Case[runArgs, string, string]{
		{
			Name: "OK closed normally",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "channel closed normally",
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints2,
				},
			},
			Invoker: runInvoker,
			Want:    ints2,
		},
		{
			Name: "NG wrong Want, closed normally",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "channel closed normally",
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints1,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  notEqualsMsg(counterInvoker.Name, ints2, ints1),
		},
		{
			Name: "OK closed by context",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name:    "channel closed by context",
					Context: ContextWithCountCancel(1),
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints1,
				},
			},
			Invoker: runInvoker,
			Want:    ints1,
		},
		{
			Name: "NG wrong Want, closed by context",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name:    "channel closed by context",
					Context: ContextWithCountCancel(1),
					Args: counterArgs{
						end: 2,
					},
					Invoker: counterInvoker,
					Want:    ints2,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  notEqualsMsg(counterInvoker.Name, ints1, ints2),
		},
		{
			Name: "OK error",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = 0",
					Args: counterArgs{
						end: 0,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgOK,
				},
			},
			Invoker: runInvoker,
		},
		{
			Name: "NG wrong error",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = 0",
					Args: counterArgs{
						end: 0,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgNG,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  wrongErrorMsg(counterInvoker.Name, errMsgOK, errMsgNG),
		},
		{
			Name: "NG not error",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = 1",
					Args: counterArgs{
						end: 1,
					},
					Invoker: counterInvoker,
					ErrMsg:  errMsgOK,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  notErrorMsg(counterInvoker.Name, errMsgOK),
		},

		{
			Name: "OK panic",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = -1",
					Args: counterArgs{
						end: -1,
					},
					Invoker: counterInvoker,
					Panic:   panicOK,
				},
			},
			Invoker: runInvoker,
		},
		{
			Name: "NG wrong panic",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = -1",
					Args: counterArgs{
						end: -1,
					},
					Invoker: counterInvoker,
					Panic:   panicNG,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  wrongPanicMsg(counterInvoker.Name, panicOK, panicNG),
		},
		{
			Name: "NG not panic",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "end = 1",
					Args: counterArgs{
						end: 1,
					},
					Invoker: counterInvoker,
					Panic:   panicOK,
				},
			},
			Invoker: runInvoker,
			ErrMsg:  notPanicMsg(counterInvoker.Name, panicOK),
		},
		{
			Name: "PANIC invoke",
			Args: runArgs{
				tc: Case[counterArgs, []int, string]{
					Name: "channel closed normally",
					Args: counterArgs{
						end: 2,
					},
					Invoker: Invoker[counterArgs, []int]{
						Name:   "Invoke is nil",
						Invoke: nil,
					},
					Want: ints2,
				},
			},
			Invoker: runInvoker,
			Panic:   "c.Invoker.Invoke must not be nil",
		},
	}

	// simple post processor that just returns the arguments as they are
	pp := func(ctx context.Context, r string, err error) (string, error) {
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
