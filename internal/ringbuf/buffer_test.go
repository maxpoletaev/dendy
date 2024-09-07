package ringbuf

import (
	"testing"

	"github.com/maxpoletaev/dendy/internal/testutil"
)

func TestBuffer_At(t *testing.T) {
	q := New[int](3)
	q.PushBack(1, 2, 3)

	testutil.Equal(t, q.At(0), 1)
	testutil.Equal(t, q.At(1), 2)
	testutil.Equal(t, q.At(2), 3)

	testutil.Panic(t, func() {
		q.At(3) // out of range
	})
}

func TestBuffer_FrontBack(t *testing.T) {
	q := New[int](3)
	q.PushBack(1)

	testutil.Equal(t, q.Front(), 1)
	testutil.Equal(t, q.Back(), 1)

	q.PushBack(2)
	testutil.Equal(t, q.Front(), 1)
	testutil.Equal(t, q.Back(), 2)

	q.PushBack(3)
	testutil.Equal(t, q.Front(), 1)
	testutil.Equal(t, q.Back(), 3)

	q.PopFront()
	testutil.Equal(t, q.Front(), 2)
	testutil.Equal(t, q.Back(), 3)
}

func TestBuffer_PushPop(t *testing.T) {
	q := New[int](3)

	q.PushBack(1)
	q.PushBack(2)
	q.PushBack(3)

	testutil.Panic(t, func() {
		q.PushBack(4) // no more space
	})

	testutil.Equal(t, q.Len(), 3)
	testutil.Equal(t, q.PopFront(), 1)
	testutil.Equal(t, q.PopBack(), 3)
	testutil.Equal(t, q.PopFront(), 2)

	testutil.Panic(t, func() {
		q.PopFront() // no more items
	})
}

func TestBuffer_TruncFront(t *testing.T) {
	q := New[int](5)

	q.PushBack(1, 2, 3, 4, 5)
	q.TruncFront(2)

	testutil.Equal(t, q.Len(), 3)
	testutil.Equal(t, q.At(0), 3)
	testutil.Equal(t, q.At(1), 4)
	testutil.Equal(t, q.At(2), 5)

	q.PushBack(6, 7)
	q.TruncFront(3)

	testutil.Equal(t, q.Len(), 2)
	testutil.Equal(t, q.At(0), 6)
	testutil.Equal(t, q.At(1), 7)
}
