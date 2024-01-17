package ringbuf

type Buffer[T any] struct {
	EmptyValue T

	items    []T
	capacity int
	length   int
	head     int
	tail     int
}

func New[T any](capacity int) *Buffer[T] {
	return &Buffer[T]{
		items:    make([]T, capacity),
		capacity: capacity,
	}
}

func (q *Buffer[T]) PushBack(items ...T) {
	if q.length+len(items) > q.capacity {
		panic("queue is full")
	}

	for _, item := range items {
		q.items[q.tail] = item
		q.tail = (q.tail + 1) % len(q.items)
		q.length++
	}
}

func (q *Buffer[T]) PopFront() T {
	if q.length == 0 {
		panic("queue is empty")
	}

	item := q.items[q.head]
	q.items[q.head] = q.EmptyValue
	q.head = (q.head + 1) % len(q.items)
	q.length--

	return item
}

func (q *Buffer[T]) TruncFront(n int) {
	if n < 0 || n > q.length {
		panic("index out of range")
	}

	for i := 0; i < n; i++ {
		q.items[(q.head+i)%len(q.items)] = q.EmptyValue
	}

	q.length -= n
	q.head = (q.head + n) % len(q.items)
	q.tail = (q.head + q.length) % len(q.items)
}

func (q *Buffer[T]) TruncBack(n int) {
	if n < 0 || n > q.length {
		panic("index out of range")
	}

	for i := 0; i < n; i++ {
		idx := (q.tail - 1 - i + len(q.items)) % len(q.items)
		q.items[idx] = q.EmptyValue
	}

	q.length -= n
	q.tail = (q.head + q.length) % len(q.items)
}

func (q *Buffer[T]) Front() T {
	if q.length == 0 {
		panic("queue is empty")
	}

	return q.items[q.head]
}

func (q *Buffer[T]) Back() T {
	if q.length == 0 {
		panic("queue is empty")
	}

	idx := (q.tail - 1 + len(q.items)) % len(q.items)

	return q.items[idx]
}

func (q *Buffer[T]) At(idx int) T {
	if idx < 0 || idx >= q.length {
		panic("index out of range")
	}

	return q.items[(q.head+idx)%len(q.items)]
}

func (q *Buffer[T]) Set(idx int, item T) {
	if idx < 0 || idx >= q.length {
		panic("index out of range")
	}

	q.items[(q.head+idx)%len(q.items)] = item
}

func (q *Buffer[T]) Len() int {
	return q.length
}

func (q *Buffer[T]) Cap() int {
	return q.capacity
}

func (q *Buffer[T]) Empty() bool {
	return q.length == 0
}

func (q *Buffer[T]) Full() bool {
	return q.length == q.capacity
}

func (q *Buffer[T]) Clear() {
	q.head = 0
	q.tail = 0
	q.length = 0
}
