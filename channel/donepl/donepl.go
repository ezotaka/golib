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
package donepl

import (
	"time"

	"github.com/ezotaka/golib/channel/ctxpl"
	"github.com/ezotaka/golib/ezctx"
)

// return channel which is closed when channel or done is closed
func OrDone[D any, T any](
	done <-chan D,
	channel <-chan T,
) <-chan T {
	return ctxpl.OrDone(ezctx.WithDone(done), channel)
}

func Repeat[D any, T any](done <-chan D, values ...T) <-chan T {
	return ctxpl.Repeat(ezctx.WithDone(done), values...)
}

func RepeatFunc[D any, T any](
	done <-chan D,
	fn func() T,
) <-chan T {
	return ctxpl.RepeatFunc(ezctx.WithDone(done), fn)
}

func Take[D any, T any](
	done <-chan D,
	valueChan <-chan T,
	num int,
) <-chan T {
	return ctxpl.Take(ezctx.WithDone(done), valueChan, num)
}

func Sleep[D any, T any](
	done <-chan D,
	c <-chan T,
	t time.Duration,
) <-chan T {
	return ctxpl.Sleep(ezctx.WithDone(done), c, t)
}

// Split the channel into two channels
func Tee[D any, T any](
	done <-chan D,
	in <-chan T,
) (<-chan T, <-chan T) {
	return ctxpl.Tee(ezctx.WithDone(done), in)
}
