package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestPushFront(t *testing.T) {
	l := NewList()
	require.Nil(t, l.Front())
	require.Nil(t, l.Back())
	require.Equal(t, 0, l.Len())

	item1 := l.PushFront(1)
	require.Equal(t, item1, l.Front())
	require.Equal(t, item1, l.Back())
	require.Equal(t, 1, l.Len())

	item2 := l.PushFront(2)
	require.Equal(t, item2, l.Front())
	require.Equal(t, item1, l.Back())
	require.Equal(t, 2, l.Len())
	require.Equal(t, item1, l.Front().Next)
	require.Nil(t, l.Front().Prev)
	require.Equal(t, item2, l.Back().Prev)
	require.Nil(t, l.Back().Next)
}

func TestPushBack(t *testing.T) {
	l := NewList()
	item1 := l.PushBack(1)
	require.Equal(t, item1, l.Front())
	require.Equal(t, item1, l.Back())
	require.Equal(t, 1, l.Len())

	item2 := l.PushBack(2)
	require.Equal(t, item1, l.Front())
	require.Equal(t, item2, l.Back())
	require.Equal(t, 2, l.Len())
	require.Equal(t, item2, l.Front().Next)
	require.Nil(t, l.Front().Prev)
	require.Equal(t, item1, l.Back().Prev)
	require.Nil(t, l.Back().Next)
}

func TestListRemove(t *testing.T) {
	t.Run("remove only element", func(t *testing.T) {
		l := NewList()
		item := l.PushFront(42)
		require.Equal(t, 1, l.Len())

		l.Remove(item)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("remove first element from many", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(1)
		second := l.PushBack(2)
		third := l.PushBack(3)

		l.Remove(first)
		require.Equal(t, 2, l.Len())
		require.Equal(t, second, l.Front())
		require.Equal(t, third, l.Back())
		require.Nil(t, second.Prev)
		require.Equal(t, third, second.Next)
		require.Equal(t, second, third.Prev)
		require.Nil(t, third.Next)
	})

	t.Run("remove last element from many", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(1)
		second := l.PushBack(2)
		third := l.PushBack(3)

		l.Remove(third)
		require.Equal(t, 2, l.Len())
		require.Equal(t, first, l.Front())
		require.Equal(t, second, l.Back())
		require.Nil(t, first.Prev)
		require.Equal(t, second, first.Next)
		require.Equal(t, first, second.Prev)
		require.Nil(t, second.Next)
	})

	t.Run("remove middle element", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(1)
		middle := l.PushBack(2)
		last := l.PushBack(3)

		l.Remove(middle)
		require.Equal(t, 2, l.Len())
		require.Equal(t, first, l.Front())
		require.Equal(t, last, l.Back())
		require.Equal(t, last, first.Next)
		require.Equal(t, first, last.Prev)
		require.Nil(t, first.Prev)
		require.Nil(t, last.Next)
	})
}

func TestList_MoveToFront(t *testing.T) {
	t.Run("move nil does nothing", func(t *testing.T) {
		l := NewList()
		l.PushBack(1)
		l.MoveToFront(nil) // не должно паниковать
		require.Equal(t, 1, l.Len())
	})

	t.Run("move front element stays front", func(t *testing.T) {
		l := NewList()
		front := l.PushBack(1)
		l.PushBack(2)

		l.MoveToFront(front)
		require.Equal(t, front, l.Front())
		require.Equal(t, 2, l.Len())
	})

	t.Run("move last element to front", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(1)
		second := l.PushBack(2)
		last := l.PushBack(3)

		l.MoveToFront(last)
		require.Equal(t, last, l.Front())
		require.Equal(t, first, l.Back())
		require.Equal(t, first, last.Next)
		require.Equal(t, last, first.Prev)
		require.Nil(t, last.Prev)
		require.Equal(t, second, first.Next)
		require.Equal(t, first, second.Prev)
		require.Nil(t, second.Next)
		require.Equal(t, 3, l.Len())
	})

	t.Run("move middle element to front", func(t *testing.T) {
		l := NewList()
		first := l.PushBack(1)
		middle := l.PushBack(2)
		last := l.PushBack(3)

		l.MoveToFront(middle)
		require.Equal(t, middle, l.Front())
		require.Equal(t, last, l.Back())
		require.Nil(t, middle.Prev)
		require.Equal(t, first, middle.Next)
		require.Equal(t, middle, first.Prev)
		require.Equal(t, last, first.Next)
		require.Equal(t, first, last.Prev)
		require.Nil(t, last.Next)
		require.Equal(t, 3, l.Len())
	})

	t.Run("move element in list with single element does nothing", func(t *testing.T) {
		l := NewList()
		only := l.PushBack(1)

		l.MoveToFront(only)
		require.Equal(t, only, l.Front())
		require.Equal(t, only, l.Back())
		require.Equal(t, 1, l.Len())
		require.Nil(t, only.Prev)
		require.Nil(t, only.Next)
	})
}
