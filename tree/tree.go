package tree

import "sync"

type Node[T any] struct {
	mu       sync.RWMutex
	children []*Node[T]
	Data     T
}

func NewNode[T any](data T) *Node[T] {
	return &Node[T]{
		Data:     data,
		children: make([]*Node[T], 0),
	}
}

func (n *Node[T]) Children() []*Node[T] {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.children
}

func (n *Node[T]) AddChild(data T) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.children = append(n.children, NewNode(data))
}

func (n *Node[T]) AddChildNode(child *Node[T]) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.children = append(n.children, child)
}

// Traverse nodes
// first executes on current node then traverses deeper into tree
// its an standard pre-order traversal
func (n *Node[T]) Traverse(act func(T)) {
	n.mu.RLock()
	data := n.Data
	children := append([]*Node[T](nil), n.children...) // Copy slice
	n.mu.RUnlock()

	act(data)
	for _, child := range children {
		child.Traverse(act)
	}
}

// Post-order traversal (children before parent)
func (n *Node[T]) TraversePostOrder(act func(T)) {
	n.mu.RLock()
	data := n.Data
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	for _, child := range children {
		child.TraversePostOrder(act)
	}
	act(data)
}

// Level-order traversal (breadth-first)
func (n *Node[T]) TraverseLevelOrder(act func(T)) {
	queue := []*Node[T]{n}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		current.mu.RLock()
		data := current.Data
		children := append([]*Node[T](nil), current.children...)
		current.mu.RUnlock()

		act(data)
		queue = append(queue, children...)
	}
}

func (n *Node[T]) Where(test func(*Node[T]) bool) *Node[T] {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if !test(n) {
		return nil
	}

	r := NewNode(n.Data)
	for _, ch := range n.children {
		if filtered := ch.Where(test); filtered != nil {
			r.children = append(r.children, filtered)
		}
	}
	return r
}

func (n *Node[T]) Search(test func(*Node[T]) bool) []*Node[T] {
	n.mu.RLock()
	defer n.mu.RUnlock()

	res := make([]*Node[T], 0)
	if test(n) {
		res = append(res, n)
	}
	for _, ch := range n.children {
		res = append(res, ch.Search(test)...)
	}
	return res
}
