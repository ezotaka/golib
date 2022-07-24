package conv

import (
	"context"
	"sync"
)

// Convert values to channel
func ToChan[T any](values ...T) <-chan T {
	ch := make(chan T, len(values))
	go func() {
		defer close(ch)
		for _, v := range values {
			ch <- v
		}
	}()
	return ch
}

// Convert channel to slice synchronously
// This function is blocked until ctx is done or c is closed
func ToSlice[T any](ctx context.Context, c <-chan T) []T {
	if c == nil {
		return nil
	}
	got := []T{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				got = append(got, v)
			}
		}
	}()
	wg.Wait()
	return got
}
