package utils

// DropOne attempts to receive and discard a single value from the channel.
// If the channel is empty, it returns immediately without blocking.
func DropOne[T any](ch <-chan T) {
	select {
	case <-ch:
	default:
	}
}

// TryDropOne attempts to receive and discard a single value from the channel.
// It returns true if a value was dropped, false if the channel was empty.
// It never blocks.
func TryDropOne[T any](ch <-chan T) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
