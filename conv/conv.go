package conv

// Convert values to channel
func Chan[T any](values ...T) <-chan T {
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
// This function is blocked until c is closed
func Slice[T any](c <-chan T) []T {
	if c == nil {
		return nil
	}
	got := []T{}
	for v := range c {
		got = append(got, v)
	}
	return got
}
