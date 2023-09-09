package ringqueue

import (
	"testing"
)

func shouldPanic(t *testing.T, f func()) {
	defer func() {
		recover()
	}()

	f()
	t.Fatalf("should have panicked")
}

func TestQueue_PushPop(t *testing.T) {
	q := New[int](3)

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	shouldPanic(t, func() {
		q.Enqueue(4)
	})

	if q.Len() != 3 {
		t.Fatalf("q.Len() = %d, want 3", q.Len())
	}

	if q.Dequeue() != 1 {
		t.Fatalf("q.Dequeue() = %d, want 1", q.Dequeue())
	}

	if q.Dequeue() != 2 {
		t.Fatalf("q.Dequeue() = %d, want 2", q.Dequeue())
	}

	if q.Dequeue() != 3 {
		t.Fatalf("q.Dequeue() = %d, want 3", q.Dequeue())
	}

	shouldPanic(t, func() {
		q.Dequeue()
	})
}
