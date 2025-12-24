package tree

import (
	"reflect"
	"sort"
	"testing"
)

// A simple implementation of DependencyNode for testing.
type depNode struct {
	name         string
	deps         []string
	dependencies []string
}

func (d depNode) Name() string {
	return d.name
}

func (d depNode) Dependents() []string {
	return d.deps
}

func (d depNode) Dependencies() []string {
	return d.dependencies
}

func TestNewForest_Simple(t *testing.T) {
	nodes := []depNode{
		{name: "a"},
		{name: "b"},
	}
	forest, err := NewForest(nodes)
	if err != nil {
		t.Fatalf("NewForest() error = %v, want nil", err)
	}
	if len(forest) != 2 {
		t.Fatalf("len(forest) = %d, want 2", len(forest))
	}
	// Sort by name to have a deterministic order
	sort.Slice(forest, func(i, j int) bool {
		return forest[i].Data.Name() < forest[j].Data.Name()
	})
	if forest[0].Data.Name() != "a" || forest[1].Data.Name() != "b" {
		t.Errorf("Forest contains wrong nodes")
	}
}

func TestNewForest_WithDependencies(t *testing.T) {
	nodes := []depNode{
		{name: "a", deps: []string{"b"}},
		{name: "b"},
	}
	forest, err := NewForest(nodes)
	if err != nil {
		t.Fatalf("NewForest() error = %v, want nil", err)
	}
	if len(forest) != 1 {
		t.Fatalf("len(forest) = %d, want 1", len(forest))
	}
	root := forest[0]
	if root.Data.Name() != "a" {
		t.Errorf("Root is %s, want a", root.Data.Name())
	}
	if len(root.Children()) != 1 {
		t.Fatalf("Root should have 1 child, has %d", len(root.Children()))
	}
	if root.Children()[0].Data.Name() != "b" {
		t.Errorf("Child is %s, want b", root.Children()[0].Data.Name())
	}
}

func TestNewForest_CircularDependency(t *testing.T) {
	nodes := []depNode{
		{name: "a", deps: []string{"b"}},
		{name: "b", deps: []string{"a"}},
	}
	_, err := NewForest(nodes)
	if err == nil {
		t.Fatal("Expected a circular dependency error, but got nil")
	}
}

func TestNewForest_Complex(t *testing.T) {
	nodes := []depNode{
		{name: "a", deps: []string{"b", "c"}},
		{name: "b", deps: []string{"d"}},
		{name: "c", deps: []string{"d"}},
		{name: "d"},
		{name: "e", deps: []string{"f"}},
		{name: "f"},
	}
	forest, err := NewForest(nodes)
	if err != nil {
		t.Fatalf("NewForest() error = %v, want nil", err)
	}
	if len(forest) != 2 {
		t.Fatalf("len(forest) = %d, want 2", len(forest))
	}

	// Sort by name for deterministic check
	sort.Slice(forest, func(i, j int) bool {
		return forest[i].Data.Name() < forest[j].Data.Name()
	})

	// Helper to count unique nodes in a DAG
	countUniqNodes := func(node *Node[depNode]) int {
		q := []*Node[depNode]{node}
		visited := make(map[*Node[depNode]]struct{})
		visited[node] = struct{}{}
		count := 0
		for len(q) > 0 {
			curr := q[0]
			q = q[1:]
			count++
			for _, child := range curr.Children() {
				if _, ok := visited[child]; !ok {
					visited[child] = struct{}{}
					q = append(q, child)
				}
			}
		}
		return count
	}

	// Check tree "a"
	if forest[0].Data.Name() != "a" {
		t.Errorf("Expected root 'a', got '%s'", forest[0].Data.Name())
	}
	if size := countUniqNodes(forest[0]); size != 4 {
		t.Errorf("Size of tree 'a' is %d, want 4", size)
	}

	// Check tree "e"
	if forest[1].Data.Name() != "e" {
		t.Errorf("Expected root 'e', got '%s'", forest[1].Data.Name())
	}
	if forest[1].Size() != 2 {
		t.Errorf("Size of tree 'e' is %d, want 2", forest[1].Size())
	}
}

func TestNewForest_UsingDependenciesField(t *testing.T) {
	nodes := []depNode{
		{name: "a"},
		{name: "b", dependencies: []string{"a"}},
	}
	forest, err := NewForest(nodes)
	if err != nil {
		t.Fatalf("NewForest() error = %v, want nil", err)
	}
	if len(forest) != 1 {
		t.Fatalf("len(forest) = %d, want 1", len(forest))
	}
	root := forest[0]
	if root.Data.Name() != "a" {
		t.Errorf("Root is %s, want a", root.Data.Name())
	}
	if len(root.Children()) != 1 {
		t.Fatalf("Root should have 1 child, has %d", len(root.Children()))
	}
	if root.Children()[0].Data.Name() != "b" {
		t.Errorf("Child is %s, want b", root.Children()[0].Data.Name())
	}
}


func TestShakeForest(t *testing.T) {
	nodes := []depNode{
		{name: "a", deps: []string{"b"}}, // size 2
		{name: "b"},
		{name: "c", deps: []string{"d"}}, // size 3
		{name: "d", deps: []string{"e"}},
		{name: "e"},
	}
	forest, _ := NewForest(nodes)

	shaken := ShakeForest(forest, func(n *Node[depNode]) bool {
		return n.Data.Name() != "e"
	})

	if len(shaken) != 2 {
		t.Fatalf("Shaken forest should have 2 trees, has %d", len(shaken))
	}

	// Sort by name for deterministic check
	sort.Slice(shaken, func(i, j int) bool {
		return shaken[i].Data.Name() < shaken[j].Data.Name()
	})

	if shaken[0].Size() != 2 {
		t.Errorf("Size of first shaken tree is %d, want 2", shaken[0].Size())
	}
	var tree1Names []string
	shaken[0].Traverse(func(d depNode) { tree1Names = append(tree1Names, d.name) })
	sort.Strings(tree1Names)
	if !reflect.DeepEqual(tree1Names, []string{"a", "b"}) {
		t.Errorf("First shaken tree has nodes %v, want ['a', 'b']", tree1Names)
	}

	if shaken[1].Size() != 2 {
		t.Errorf("Size of second shaken tree is %d, want 2", shaken[1].Size())
	}
	var tree2Names []string
	shaken[1].Traverse(func(d depNode) { tree2Names = append(tree2Names, d.name) })
	sort.Strings(tree2Names)
	if !reflect.DeepEqual(tree2Names, []string{"c", "d"}) {
		t.Errorf("Second shaken tree has nodes %v, want ['c', 'd']", tree2Names)
	}

}