package generic

type Queue[T any] struct {
	EmptyValue T

	items []T
	head  int
	tail  int
	size  int
}

func NewQueue[T any](capacity int) *Queue[T] {
	return &Queue[T]{
		items: make([]T, capacity),
	}
}

func (q *Queue[T]) Enqueue(item T) {
	if q.size == len(q.items) {
		panic("queue is full")
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % len(q.items)
	q.size++
}

func (q *Queue[T]) Dequeue() T {
	if q.size == 0 {
		panic("queue is EmptyValue")
	}

	item := q.items[q.head]
	q.items[q.head] = q.EmptyValue
	q.head = (q.head + 1) % len(q.items)
	q.size--

	return item
}

func (q *Queue[T]) Front() T {
	if q.size == 0 {
		panic("queue is EmptyValue")
	}

	return q.items[q.head]
}

func (q *Queue[T]) Len() int {
	return q.size
}

func (q *Queue[T]) Cap() int {
	return len(q.items)
}

func (q *Queue[T]) Empty() bool {
	return q.size == 0
}

func (q *Queue[T]) Full() bool {
	return q.size == len(q.items)
}

func (q *Queue[T]) Clear() {
	q.head = 0
	q.tail = 0
	q.size = 0
}
