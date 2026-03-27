package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front  *ListItem
	back   *ListItem
	length int
}

func (l *list) Front() *ListItem {
	return l.front
}

func (l *list) Back() *ListItem {
	return l.back
}

func (l *list) Len() int {
	return l.length
}

func (l *list) PushFront(v interface{}) *ListItem {
	newNode := &ListItem{
		Value: v,
		Next:  l.front,
		Prev:  nil,
	}
	if l.front != nil {
		l.front.Prev = newNode
	}
	l.front = newNode
	if l.back == nil {
		l.back = newNode
	}
	l.length++
	return newNode
}

func (l *list) PushBack(v interface{}) *ListItem {
	newNode := &ListItem{
		Value: v,
		Prev:  l.back,
		Next:  nil,
	}
	if l.back != nil {
		l.back.Next = newNode
	}
	l.back = newNode
	if l.front == nil {
		l.front = newNode
	}
	l.length++
	return newNode
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.front = i.Next
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.back = i.Prev
	}
	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || l.front == i {
		return
	}
	i.Prev.Next = i.Next
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.back = i.Prev
	}

	i.Prev = nil
	i.Next = l.front
	if l.front != nil {
		l.front.Prev = i
	}
	l.front = i
	if l.back == nil {
		l.back = i
	}
}

func NewList() List {
	return new(list)
}
