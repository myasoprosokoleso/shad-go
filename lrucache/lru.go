//go:build !solution

package lrucache

import "container/list"

type node struct {
	key   int
	value int
}

type lrucache struct {
	cap    int
	keys   map[int]*list.Element
	values *list.List
}

func New(cap int) Cache {
	return &lrucache{cap: cap, keys: make(map[int]*list.Element, cap), values: list.New()}
}

func (l *lrucache) Get(key int) (int, bool) {
	if n, ok := l.keys[key]; ok {
		l.values.MoveToFront(n)
		return n.Value.(*node).value, true
	}
	return 0, false
}

func (l *lrucache) Set(key, value int) {
	if l.cap <= 0 {
		return
	}

	if n, ok := l.keys[key]; ok {
		n.Value.(*node).value = value
		l.values.MoveToFront(n)
		return
	}

	if len(l.keys) == l.cap {
		back := l.values.Back()
		if back != nil {
			l.values.Remove(back)
			delete(l.keys, back.Value.(*node).key)
		}
	}

	front := l.values.PushFront(&node{key: key, value: value})
	l.keys[key] = front
}

func (l *lrucache) Range(f func(key, value int) bool) {
	for n := l.values.Back(); n != nil; n = n.Prev() {
		v := n.Value.(*node)
		if ok := f(v.key, v.value); !ok {
			return
		}
	}
}

func (l *lrucache) Clear() {
	for n := l.values.Front(); n != nil; {
		next := n.Next()
		l.values.Remove(n)
		n = next
	}

	clear(l.keys)
}
