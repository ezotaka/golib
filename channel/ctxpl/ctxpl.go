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

// TODO: Swap the relationship between donepl and ctxpl
package ctxpl

import (
	"context"
	"time"
)

// return channel which is closed when channel or done is closed
func OrDone[T any](
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

func Repeat[T any](ctx context.Context, values ...T) <-chan T {
	valuesChan := make(chan T)
	select {
	case <-ctx.Done():
		close(valuesChan)
	default:
		go func() {
			defer close(valuesChan)
			if len(values) == 0 {
				return
			}
			for {
				for _, v := range values {
					select {
					case <-ctx.Done():
						return
					case valuesChan <- v:
					}
				}
			}
		}()
	}
	return valuesChan
}

func RepeatFunc[T any](
	ctx context.Context,
	fn func() T,
) <-chan T {
	if fn == nil {
		panic("fn must not be nil")
	}
	valueChan := make(chan T)
	go func() {
		defer close(valueChan)
		for {
			select {
			case <-ctx.Done():
				return
			case valueChan <- fn():
			}
		}
	}()
	return valueChan
}

func Take[T any](
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

func Sleep[T any](
	ctx context.Context,
	c <-chan T,
	t time.Duration,
) <-chan T {
	if t == 0 {
		return c
	}
	ch := make(chan T)
	go func() {
		defer close(ch)
		p := OrDone(ctx, c)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-p:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
				case <-time.After(t):
					ch <- v
				}
			}
		}
	}()
	return ch
}

// Split the channel into two channels
func Tee[T any](
	ctx context.Context,
	in <-chan T,
) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)
	go func() {
		defer close(out1)
		defer close(out2)
		for val := range OrDone(ctx, in) {
			var out1, out2 = out1, out2
			// Writes reliably to two channels
			for i := 0; i < 2; i++ {
				select {
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}
