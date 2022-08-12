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

package channel

import (
	"context"

	"github.com/ezotaka/golib/conv"
	"github.com/ezotaka/golib/eztest"
)

// TODO: copy and paste from ctxpl package
// return channel which is closed when channel or done is closed
func orDone[T any](
	ctx context.Context,
	channel <-chan T,
) <-chan T {
	valChan := make(chan T)
	go func() {
		defer close(valChan)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-channel:
				if !ok {
					return
				}
				select {
				case valChan <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return valChan
}

func take[T any](
	ctx context.Context,
	valueChan <-chan T,
	num int,
) <-chan T {
	if valueChan == nil {
		return nil
	}
	takeChan := make(chan T)
	go func() {
		defer close(takeChan)
		for i := 0; i < num; i++ {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-valueChan:
				if !ok {
					return
				}
				takeChan <- v
			}
		}
	}()
	return takeChan
}

// return channel which can be cancelled by context
func withCountCancel[T any](ctx context.Context, c <-chan T) <-chan T {
	if cnt, ok := eztest.CountToCancel(ctx); ok {
		return orDone(ctx, take(ctx, c, cnt))
	} else {
		return orDone(ctx, c)
	}
}

// RunTest channel test using test case defined by Case
func RunTest[
	// Type of args which is passed to the function to be tested
	A any,
	// Type of chanel which is returned by the function to be tested
	C any,
](
	// Test case for the function to be tested
	tc eztest.Case[A, <-chan C, []C],
) ([]C, error) {
	// [post process]
	// Channel c which is returned by the function to be tested can be canceled by the context.
	// Synchronously converts the value sent from the channel into slices.
	pp := func(ctx context.Context, c <-chan C, err error) ([]C, error) {
		if c == nil || err != nil {
			return nil, err
		}
		return conv.Slice(withCountCancel(ctx, c)), nil
	}

	return eztest.Run(tc, pp)
}
