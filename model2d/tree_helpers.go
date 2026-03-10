package model2d

import "github.com/unixpickle/splaytree"

// splayTreeIterateBackwards iterates backwards before an element of the tree,
// wrapping until but not including the original value.
func splayTreeIterateBackwards[T splaytree.Value[T]](t *splaytree.Tree[T], v T, f func(T) bool) {
	if !splayTreeIterateBefore(t, v, f) {
		return
	}
	splayTreeIterateFromEnd(t, func(value T) bool {
		if value.Compare(v) <= 0 {
			return false
		}
		return f(value)
	})
}

// splayTreeIterateForwards iterates forwards after an element of the tree,
// wrapping until but not including the original value.
func splayTreeIterateForwards[T splaytree.Value[T]](t *splaytree.Tree[T], v T, f func(T) bool) {
	if !splayTreeIterateAfter(t, v, f) {
		return
	}
	t.Iterate(func(value T) bool {
		if value.Compare(v) >= 0 {
			return false
		}
		return f(value)
	})
}

func splayTreeIterateFromEnd[T splaytree.Value[T]](t *splaytree.Tree[T], f func(T) bool) bool {
	var itFn func(n *splaytree.Node[T]) bool
	itFn = func(n *splaytree.Node[T]) bool {
		if n == nil {
			return true
		}
		if !itFn(n.Right) {
			return false
		}
		if !f(n.Value) {
			return false
		}
		return itFn(n.Left)
	}
	return itFn(t.Root)
}

func splayTreeIterateBefore[T splaytree.Value[T]](t *splaytree.Tree[T], v T, f func(T) bool) bool {
	var itFn func(n *splaytree.Node[T]) bool
	itFn = func(n *splaytree.Node[T]) bool {
		if n == nil {
			return true
		}
		cmp := n.Value.Compare(v)
		if cmp >= 0 {
			return itFn(n.Left)
		} else {
			if !itFn(n.Right) {
				return false
			}
			if !f(n.Value) {
				return false
			}
			return itFn(n.Left)
		}
	}
	return itFn(t.Root)
}

func splayTreeIterateAfter[T splaytree.Value[T]](t *splaytree.Tree[T], v T, f func(T) bool) bool {
	var itFn func(n *splaytree.Node[T]) bool
	itFn = func(n *splaytree.Node[T]) bool {
		if n == nil {
			return true
		}
		cmp := n.Value.Compare(v)
		if cmp <= 0 {
			return itFn(n.Right)
		} else {
			if !itFn(n.Left) {
				return false
			}
			if !f(n.Value) {
				return false
			}
			return itFn(n.Right)
		}
	}
	return itFn(t.Root)
}
