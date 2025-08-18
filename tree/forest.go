package tree

import "fmt"

type Forest[T any] = []*Node[T]

type DependencyNode[T comparable] interface {
	Name() T
	Dependents() []T
	Dependencies() []T
}

// Create new trees based on given Dependency Nodes
// Can Detect Circular Dependency
// Creates new tree if a node has no dependent
// A single node may appear on other trees as well, thus output node count
// may be far more than what input was
func NewForest[R comparable, T DependencyNode[R]](nodes []T) (Forest[T], error) {
	nodeMap := make(map[R]T)
	for _, node := range nodes {
		nodeMap[node.Name()] = node
	}

	// Step 2: Build complete dependency graph from both dependencies and dependents
	dependencies := make(map[R][]R) // node -> its dependencies

	// Process dependencies field
	for _, node := range nodes {
		name := node.Name()
		if _, exists := dependencies[name]; !exists {
			dependencies[name] = []R{}
		}

		dependencies[name] = append(dependencies[name], node.Dependencies()...)

	}

	// Process dependents field (inverse relationship)
	for _, node := range nodes {
		name := node.Name()
		for _, dependent := range node.Dependents() {
			if _, exists := dependencies[dependent]; !exists {
				dependencies[dependent] = []R{}
			}
			// If A has B as dependent, then B depends on A
			dependencies[dependent] = append(dependencies[dependent], name)
		}
	}

	// Step 3: Build adjacency list and calculate in-degrees
	inDegree := make(map[R]int)
	adjList := make(map[R][]R) // node -> nodes that depend on it

	for _, node := range nodes {
		name := node.Name()
		if _, exists := inDegree[name]; !exists {
			inDegree[name] = 0
		}
		if _, exists := adjList[name]; !exists {
			adjList[name] = []R{}
		}
	}

	for node, deps := range dependencies {
		inDegree[node] = len(deps)
		for _, dep := range deps {
			adjList[dep] = append(adjList[dep], node)
		}
	}

	// Step 4: Find all nodes with no dependencies (roots)
	queue := []R{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Step 5: Process nodes using Kahn's algorithm with duplicate prevention
	processed := 0
	nodeTree := make(map[R]*Node[T])
	roots := []*Node[T]{}
	childSet := make(map[R]map[R]struct{}) // parent -> set of child names

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed++

		// Create node if not exists
		if _, exists := nodeTree[current]; !exists {
			nodeTree[current] = NewNode(nodeMap[current])
			// If this node has no dependencies, it's a root
			if inDegree[current] == 0 {
				roots = append(roots, nodeTree[current])
			}
			// Initialize child set for this node
			childSet[current] = make(map[R]struct{})
		}

		// Process nodes that depend on current
		for _, dependent := range adjList[current] {
			inDegree[dependent]--

			// Create dependent node if not exists
			if _, exists := nodeTree[dependent]; !exists {
				nodeTree[dependent] = NewNode(nodeMap[dependent])
				childSet[dependent] = make(map[R]struct{})
			}

			// Only add child if not already present
			if _, exists := childSet[current][dependent]; !exists {
				nodeTree[current].children = append(nodeTree[current].children, nodeTree[dependent])
				childSet[current][dependent] = struct{}{}
			}

			// If all dependencies resolved, add to queue
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Step 6: Check for circular dependencies
	if processed != len(nodes) {
		// Find nodes involved in cycles
		cycleNodes := []R{}
		for name := range inDegree {
			if inDegree[name] > 0 {
				cycleNodes = append(cycleNodes, name)
			}
		}
		return nil, fmt.Errorf("circular dependency detected involving nodes: %v", cycleNodes)
	}

	return roots, nil
}

// ShakeForest calls `Shake` on each tree of the forest
//   - Performs way worse than `Shake` because of possibility of entangled branches on multiple trees.
func ShakeForest[T any](f Forest[T], test func(*Node[T]) bool) Forest[T] {
	res := make(Forest[T], 0, len(f))
	for _, t := range f {
		nt, ok := t.Shake(test)
		if ok {
			res = append(res, nt)
		}
	}
	return res
}
