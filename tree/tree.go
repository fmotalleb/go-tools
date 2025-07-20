package tree

type Node[T any] struct {
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
	return n.children
}

func (n *Node[T]) AddChild(data T) {
	n.children = append(n.children, NewNode(data))
}

func (n *Node[T]) Traverse(act func(T)) {
	act(n.Data)
	for _, child := range n.children {
		child.Traverse(act)
	}
}

// Post-order traversal (children before parent)
func (n *Node[T]) TraversePostOrder(act func(T)) {
	for _, child := range n.children {
		child.TraversePostOrder(act)
	}
	act(n.Data)
}

// Level-order traversal (breadth-first)
func (n *Node[T]) TraverseLevelOrder(act func(T)) {
	queue := []*Node[T]{n}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		act(current.Data)

		queue = append(queue, current.children...)
	}
}
