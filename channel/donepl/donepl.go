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
