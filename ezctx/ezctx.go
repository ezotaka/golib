package ezctx

import (
	"context"
)

// Return context cancelled when done channel is closed
func WithDone[T any](done <-chan T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	if done == nil {
		cancel()
	} else {
		select {
		case <-done:
			cancel()
		default:
			go func() {
				<-done
				cancel()
			}()
		}
	}
	return ctx
}
