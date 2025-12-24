package tree

import (
	"reflect"
	"sort"
	"testing"
)

// Helper function to create a simple tree for testing.
func createTestTree() *Node[int] {
	root := NewNode(1)
	root.AddChild(2)
	root.AddChild(3)
	root.Children()[0].AddChild(4)
	root.Children()[0].AddChild(5)
	root.Children()[1].AddChild(6)
	return root
}

func TestNewNode(t *testing.T) {
	node := NewNode(10)
	if node.Data != 10 {
		t.Errorf("Expected data to be 10, got %d", node.Data)
	}
	if len(node.children) != 0 {
		t.Errorf("Expected children to be empty, but got %d", len(node.children))
	}
}

func TestAddChild(t *testing.T) {
	root := NewNode(1)
	root.AddChild(2)
	if len(root.children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.children))
	}
	if root.children[0].Data != 2 {
		t.Errorf("Expected child data to be 2, got %d", root.children[0].Data)
	}
}

func TestAddChildNode(t *testing.T) {
	root := NewNode(1)
	child := NewNode(2)
	root.AddChildNode(child)
	if len(root.children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(root.children))
	}
	if root.children[0] != child {
		t.Error("Child node was not added correctly")
	}
}

func TestChildren(t *testing.T) {
	root := NewNode(1)
	root.AddChild(2)
	root.AddChild(3)
	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(children))
	}
	if children[0].Data != 2 || children[1].Data != 3 {
		t.Error("Children data is incorrect")
	}
	// Modify the returned slice to ensure it's a copy
	children[0] = NewNode(99)
	if root.children[0].Data == 99 {
		t.Error("Children slice is not a copy")
	}
}

func TestTraverse(t *testing.T) {
	tree := createTestTree()
	var result []int
	tree.Traverse(func(i int) {
		result = append(result, i)
	})
	expected := []int{1, 2, 4, 5, 3, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Traverse() = %v, want %v", result, expected)
	}
}

func TestTraverseNode(t *testing.T) {
	tree := createTestTree()
	var result []int
	tree.TraverseNode(func(n *Node[int]) {
		result = append(result, n.Data)
	})
	expected := []int{1, 2, 4, 5, 3, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TraverseNode() = %v, want %v", result, expected)
	}
}

func TestTraversePostOrder(t *testing.T) {
	tree := createTestTree()
	var result []int
	tree.TraversePostOrder(func(i int) {
		result = append(result, i)
	})
	expected := []int{4, 5, 2, 6, 3, 1}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TraversePostOrder() = %v, want %v", result, expected)
	}
}

func TestTraverseLevelOrder(t *testing.T) {
	tree := createTestTree()
	var result []int
	tree.TraverseLevelOrder(func(i int) {
		result = append(result, i)
	})
	expected := []int{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TraverseLevelOrder() = %v, want %v", result, expected)
	}
}

func TestWhere(t *testing.T) {
	tree := createTestTree()
	filtered, ok := tree.Where(func(n *Node[int]) bool {
		return n.Data%2 == 0
	})
	if ok {
		t.Error("Expected root to be filtered out, but it wasn't")
	}

	tree2 := createTestTree()
	filtered, ok = tree2.Where(func(n *Node[int]) bool {
		return n.Data < 5
	})

	if !ok {
		t.Fatal("Expected filtering to be successful, but it was not")
	}

	var result []int
	filtered.Traverse(func(i int) {
		result = append(result, i)
	})
	expected := []int{1, 2, 4, 3}
	sort.Ints(result)
	sort.Ints(expected)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Where() result = %v, want %v", result, expected)
	}
}

func TestSearch(t *testing.T) {
	tree := createTestTree()
	nodes := tree.Search(func(n *Node[int]) bool {
		return n.Data > 3
	})
	var result []int
	for _, n := range nodes {
		result = append(result, n.Data)
	}
	sort.Ints(result)
	expected := []int{4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Search() = %v, want %v", result, expected)
	}
}

func TestDepthMax(t *testing.T) {
	tree := createTestTree()
	if depth := tree.DepthMax(); depth != 3 {
		t.Errorf("DepthMax() = %d, want 3", depth)
	}
	if depth := tree.Children()[1].DepthMax(); depth != 2 {
		t.Errorf("DepthMax() on child = %d, want 2", depth)
	}
}

func TestSize(t *testing.T) {
	tree := createTestTree()
	if size := tree.Size(); size != 6 {
		t.Errorf("Size() = %d, want 6", size)
	}
	if size := tree.Children()[0].Size(); size != 3 {
		t.Errorf("Size() on child = %d, want 3", size)
	}
}

func TestShake(t *testing.T) {
	root := NewNode(1)
	c1 := NewNode(2)
	c2 := NewNode(3)
	c1_1 := NewNode(4)
	c1_1_1 := NewNode(5) // This should be shaken off

	root.AddChildNode(c1)
	root.AddChildNode(c2)
	c1.AddChildNode(c1_1)
	c1_1.AddChildNode(c1_1_1)

	shaken, ok := root.Shake(func(n *Node[int]) bool {
		// Keep nodes with data < 5 and their parents
		return n.Data < 5
	})

	if !ok {
		t.Fatal("Shake failed")
	}

	if size := shaken.Size(); size != 4 {
		t.Errorf("Shaken tree size = %d, want 4", size)
	}

	var result []int
	shaken.Traverse(func(i int) {
		result = append(result, i)
	})
	expected := []int{1, 2, 3, 4}
	sort.Ints(result)
	sort.Ints(expected)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Shaken tree data = %v, want %v", result, expected)
	}
}
