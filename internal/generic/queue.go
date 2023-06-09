package generic

import (
	"container/list"
)

type Queue[T any] struct {
	list *list.List
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{list: list.New()}
}

func (q *Queue[T]) Len() int {
	return q.list.Len()
}

func (q *Queue[T]) Push(v T) {
	q.list.PushBack(v)
}

func (q *Queue[T]) Pop() (v T, ok bool) {
	e := q.list.Front()
	if e == nil {
		return
	}
	q.list.Remove(e)
	v = e.Value.(T)
	ok = true
	return
}

func (q *Queue[T]) Peek() (v T, ok bool) {
	e := q.list.Front()
	if e == nil {
		return
	}
	v = e.Value.(T)
	ok = true
	return
}
