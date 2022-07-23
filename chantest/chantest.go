package chantest

import (
	"context"

	"github.com/ezotaka/golib/conv"
	"github.com/ezotaka/golib/eztest"
)

// Run channel test using test case defined by Case
func Run[
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
		if err != nil {
			return nil, err
		}
		endedChan := make(chan C)
		go func() {
			defer close(endedChan)

			if cnt, ok := eztest.CountToCancel(ctx); ok && cnt == 0 {
				return
			}

			count := 1
			for {
				select {
				case <-ctx.Done():
					return
				default:
					select {
					case val, ok := <-c:
						if !ok {
							return
						}
						endedChan <- val
						if cnt, ok := eztest.CountToCancel(ctx); ok {
							if cnt == count {
								return
							}
						}
						count++
					default:
					}
				}
			}
		}()
		return conv.ToSlice(context.Background(), endedChan), nil
	}

	return eztest.Run(tc, pp)
}
