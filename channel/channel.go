package channel

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

func ToChan[T any](done <-chan any, values ...T) <-chan T {
	ch := make(chan T, len(values))
	go func() {
		defer close(ch)
		for _, v := range values {
			select {
			case <-done:
				return
			case ch <- v:
			}
		}
	}()
	return ch
}

func Repeat[T any](done <-chan any, values ...T) <-chan T {
	valuesChan := make(chan T)
	go func() {
		defer close(valuesChan)
		if len(values) == 0 {
			return
		}
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valuesChan <- v:
				}
			}
		}
	}()
	return valuesChan
}

func RepeatFunc[T any](
	done <-chan any,
	fn func() T,
) <-chan T {
	valueChan := make(chan T)
	go func() {
		defer close(valueChan)
		if fn == nil {
			return
		}
		for {
			select {
			case <-done:
				return
			case valueChan <- fn():
			}
		}
	}()
	return valueChan
}

func Take[T any](
	done <-chan any,
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
			case <-done:
				return
			case takeChan <- <-valueChan:
			}
		}
	}()
	return takeChan
}
