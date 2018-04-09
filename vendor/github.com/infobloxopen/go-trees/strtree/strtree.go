// Package strtree implements red-black tree for key value pairs with string keys and custom comparison.
package strtree

import "strings"

// Compare defines function interface for custom comparison. Function implementing the interface should return value less than zero if its first argument precedes second one, zero if both are equal and positive if the second precedes.
type Compare func(a, b string) int

// Tree is a red-black tree for key-value pairs where key is string.
type Tree struct {
	root    *node
	compare Compare
}

// Pair is a key-value pair representing tree node content.
type Pair struct {
	Key   string
	Value interface{}
}

// NewTree creates empty tree with default comparison operation (strings.Compare).
func NewTree() *Tree {
	return &Tree{compare: strings.Compare}
}

// NewTreeWithCustomComparison creates empty tree with given comparison operation.
func NewTreeWithCustomComparison(compare Compare) *Tree {
	return &Tree{compare: compare}
}

// Insert puts given key-value pair to the tree and returns pointer to new root.
func (t *Tree) Insert(key string, value interface{}) *Tree {
	var (
		n *node
		c Compare
	)

	if t == nil {
		c = strings.Compare
	} else {
		n = t.root
		c = t.compare
	}

	return &Tree{root: n.insert(key, value, c), compare: c}
}

// InplaceInsert inserts or replaces given key-value pair in the tree. The method inserts data directly to current tree so make sure you have exclusive access to it.
func (t *Tree) InplaceInsert(key string, value interface{}) {
	t.root = t.root.inplaceInsert(key, value, t.compare)
}

// Get returns value by given key.
func (t *Tree) Get(key string) (interface{}, bool) {
	if t == nil {
		return nil, false
	}

	return t.root.get(key, t.compare)
}

// Enumerate returns channel which is populated by key pair values in order of keys.
func (t *Tree) Enumerate() chan Pair {
	ch := make(chan Pair)

	go func() {
		defer close(ch)

		if t == nil {
			return
		}

		t.root.enumerate(ch)
	}()

	return ch
}

// Delete removes node by given key. It returns copy of tree and true if node has been indeed deleted otherwise original tree and false.
func (t *Tree) Delete(key string) (*Tree, bool) {
	if t == nil {
		return nil, false
	}

	c := t.compare
	root, ok := t.root.del(key, c)
	return &Tree{root: root, compare: c}, ok
}

// IsEmpty returns true if given tree has no nodes.
func (t *Tree) IsEmpty() bool {
	return t == nil || t.root == nil
}

// Dot dumps tree to Graphviz .dot format.
func (t *Tree) Dot() string {
	body := ""

	if t != nil {
		body = t.root.dot()
	}

	return "digraph d {\n" + body + "}\n"
}
