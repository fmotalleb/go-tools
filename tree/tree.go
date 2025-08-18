package tree

import "sync"

// Node represents a concurrent-safe generic tree node.
type Node[T any] struct {
	mu       sync.RWMutex
	children []*Node[T]
	Data     T
}

// NewNode creates a new tree node with given data.
func NewNode[T any](data T) *Node[T] {
	return &Node[T]{
		Data:     data,
		children: make([]*Node[T], 0),
	}
}

// Children returns a snapshot of child nodes.
// Copies slice to avoid race conditions.
func (n *Node[T]) Children() []*Node[T] {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return append([]*Node[T](nil), n.children...)
}

// AddChild creates and adds a new child node.
func (n *Node[T]) AddChild(data T) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.children = append(n.children, NewNode(data))
}

// AddChildNode adds an existing child node.
func (n *Node[T]) AddChildNode(child *Node[T]) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.children = append(n.children, child)
}

// Traverse performs pre-order traversal (node -> children).
// Works on node data only.
func (n *Node[T]) Traverse(act func(T)) {
	n.mu.RLock()
	data := n.Data
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	act(data)
	for _, child := range children {
		child.Traverse(act)
	}
}

// TraverseNode performs pre-order traversal, exposing nodes.
func (n *Node[T]) TraverseNode(act func(*Node[T])) {
	n.mu.RLock()
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	act(n)
	for _, child := range children {
		child.TraverseNode(act)
	}
}

// TraversePostOrder performs post-order traversal (children -> node).
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

// TraverseLevelOrder performs breadth-first traversal.
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

// Where filters nodes recursively and returns a new tree.
//   - Entire branch is discarded if a parent fails the test.
func (n *Node[T]) Where(test func(*Node[T]) bool) (*Node[T], bool) {
	n.mu.RLock()
	data := n.Data
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	if !test(n) {
		return nil, false
	}

	r := NewNode(data)
	for _, ch := range children {
		if filtered, ok := ch.Where(test); ok {
			r.children = append(r.children, filtered)
		}
	}
	return r, true
}

// Search finds all nodes matching the test function.
func (n *Node[T]) Search(test func(*Node[T]) bool) []*Node[T] {
	n.mu.RLock()
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	res := make([]*Node[T], 0)
	if test(n) {
		res = append(res, n)
	}
	for _, ch := range children {
		res = append(res, ch.Search(test)...)
	}
	return res
}

// DepthMax returns maximum depth from current node to any leaf.
func (n *Node[T]) DepthMax() int {
	n.mu.RLock()
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	deepest := 0
	for _, c := range children {
		if depth := c.DepthMax(); depth > deepest {
			deepest = depth
		}
	}
	return deepest + 1
}

// Size counts nodes in the subtree rooted at this node.
func (n *Node[T]) Size() int {
	n.mu.RLock()
	children := append([]*Node[T](nil), n.children...)
	n.mu.RUnlock()

	size := 1
	for _, c := range children {
		size += c.Size()
	}
	return size
}

// Shake repeatedly prunes the tree using test until size stabilizes.
// Very expensive: uses repeated Size() + Where() calls.
func (n *Node[T]) Shake(test func(*Node[T]) bool) (*Node[T], bool) {
	prevSize := -1
	node := n
	ok := true
	var filtered *Node[T]
	// Continue until size stabilizes
	for {
		currSize := node.Size()
		if currSize == prevSize {
			break
		}
		prevSize = currSize

		filtered, ok = node.Where(test)
		if !ok {
			return nil, false
		}
		node = filtered
	}
	return node, ok
}
