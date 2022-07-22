package channel

import (
	"context"
	"time"
)

func Or[T any](channels ...<-chan T) <-chan T {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan T)
	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-Or(append(channels[3:], orDone)...):
			}
		}
	}()
	return orDone
}

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
	takeChan := make(chan T)
	go func() {
		defer close(takeChan)
		if valueChan == nil {
			return
		}
		for i := 0; i < num; i++ {
			select {
			case <-ctx.Done():
				return
			case takeChan <- <-valueChan:
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
