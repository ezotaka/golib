package channel

// Type of enumerated value in for statement
type forEnum[T any] struct {
	I int
	V T
}

// Enumerate value and index that are received from channel as forEnum struct
func Enumerate[T any](c <-chan T) <-chan forEnum[T] {
	if c == nil {
		return nil
	}
	enumChan := make(chan forEnum[T])
	go func() {
		defer close(enumChan)
		i := 0
		for v := range c {
			enumChan <- forEnum[T]{i, v}
			i++
		}
	}()
	return enumChan
}

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
