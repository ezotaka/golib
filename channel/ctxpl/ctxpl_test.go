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

package ctxpl

import (
	"context"
	"testing"
	"time"

	"github.com/ezotaka/golib/channel"
	"github.com/ezotaka/golib/conv"
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
			if _, err := channel.RunTest(tt); err != nil {
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
			if _, err := channel.RunTest(tt); err != nil {
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
			if _, err := channel.RunTest(tt); err != nil {
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
			if _, err := channel.RunTest(tt); err != nil {
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
			if _, err := channel.RunTest(tt); err != nil {
				t.Errorf(err.Error())
			}
		})
	}
}
