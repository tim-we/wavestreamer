package utils

import "sync"

// ConcurrentQueue is a thread-safe FIFO queue.
type ConcurrentQueue[T any] struct {
	list  []T
	mutex sync.RWMutex
}

// NewConcurrentQueue creates a new empty queue with the given initial capacity.
func NewConcurrentQueue[T any](capacity int) *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{
		list: make([]T, 0, capacity),
	}
}

// Add appends one or more items to the end of the queue.
func (queue *ConcurrentQueue[T]) Add(items ...T) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.list = append(queue.list, items...)
}

// Prepends adds the given items to the start of the queue.
func (queue *ConcurrentQueue[T]) Prepend(items ...T) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.list = append(items, queue.list...)
}

// Peek returns the next item without removing it.
// It returns nil if the queue is empty.
func (queue *ConcurrentQueue[T]) Peek() T {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()

	if len(queue.list) == 0 {
		var zero T
		return zero
	}
	return queue.list[0]
}

// GetNext removes and returns the next item in the queue.
// The boolean result reports whether the queue still contains items.
func (queue *ConcurrentQueue[T]) GetNext() (T, bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	var zero T

	if len(queue.list) == 0 {
		return zero, false
	}

	item := queue.list[0]
	queue.list[0] = zero // avoid retaining references
	queue.list = queue.list[1:]

	return item, true
}

// Size returns the number of items currently in the queue.
func (queue *ConcurrentQueue[T]) Size() int {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()

	return len(queue.list)
}

// IsEmpty checks if there are any more items in the queue.
func (queue *ConcurrentQueue[T]) IsEmpty() bool {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()

	return len(queue.list) == 0
}
