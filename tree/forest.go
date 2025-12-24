package tree

import (
	"fmt"
	"slices"
)

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
// may be far more than what input was.
func NewForest[R comparable, T DependencyNode[R]](nodes []T) (Forest[T], error) {
	nodeMap := indexNodes(nodes)

	dependencies := buildDependencyMap(nodes)
	inDegree, adjList := buildGraph(nodes, dependencies)

	roots, processed := buildForest(nodeMap, inDegree, adjList)

	if processed != len(nodes) {
		return nil, cycleError(inDegree)
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
	res = slices.Clip(res)
	return res
}

func indexNodes[R comparable, T DependencyNode[R]](nodes []T) map[R]T {
	m := make(map[R]T, len(nodes))
	for _, n := range nodes {
		m[n.Name()] = n
	}
	return m
}

func buildDependencyMap[R comparable, T DependencyNode[R]](nodes []T) map[R][]R {
	deps := make(map[R][]R)

	for _, n := range nodes {
		name := n.Name()
		deps[name] = append(deps[name], n.Dependencies()...)
	}

	for _, n := range nodes {
		parent := n.Name()
		for _, child := range n.Dependents() {
			deps[child] = append(deps[child], parent)
		}
	}

	return deps
}

func buildGraph[R comparable, T DependencyNode[R]](
	nodes []T,
	dependencies map[R][]R,
) (map[R]int, map[R][]R) {
	inDegree := make(map[R]int)
	adjList := make(map[R][]R)

	for _, n := range nodes {
		name := n.Name()
		inDegree[name] = 0
		adjList[name] = nil
	}

	for node, deps := range dependencies {
		inDegree[node] = len(deps)
		for _, dep := range deps {
			adjList[dep] = append(adjList[dep], node)
		}
	}

	return inDegree, adjList
}

func buildForest[R comparable, T DependencyNode[R]](
	nodeMap map[R]T,
	inDegree map[R]int,
	adjList map[R][]R,
) ([]*Node[T], int) {
	queue := make([]R, 0)
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}

	nodeTree := make(map[R]*Node[T])
	childSet := make(map[R]map[R]struct{})
	roots := make([]*Node[T], 0, len(queue))
	for _, name := range queue {
		roots = append(roots, getOrCreateNode(name, nodeMap, nodeTree, childSet))
	}
	processed := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		processed++

		currNode := getOrCreateNode(current, nodeMap, nodeTree, childSet)

		for _, dep := range adjList[current] {
			inDegree[dep]--

			depNode := getOrCreateNode(dep, nodeMap, nodeTree, childSet)

			if _, exists := childSet[current][dep]; !exists {
				currNode.children = append(currNode.children, depNode)
				childSet[current][dep] = struct{}{}
			}

			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	return roots, processed
}

func getOrCreateNode[R comparable, T DependencyNode[R]](
	name R,
	nodeMap map[R]T,
	nodeTree map[R]*Node[T],
	childSet map[R]map[R]struct{},
) *Node[T] {
	if n, ok := nodeTree[name]; ok {
		return n
	}

	n := NewNode(nodeMap[name])
	nodeTree[name] = n
	childSet[name] = make(map[R]struct{})
	return n
}

func cycleError[R comparable](inDegree map[R]int) error {
	var cycleNodes []R
	for name, deg := range inDegree {
		if deg > 0 {
			cycleNodes = append(cycleNodes, name)
		}
	}
	return fmt.Errorf("circular dependency detected involving nodes: %v", cycleNodes)
}
