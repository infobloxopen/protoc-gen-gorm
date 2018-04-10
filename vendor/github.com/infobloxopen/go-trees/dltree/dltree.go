// Package dltree implements red-black tree for key value pairs with domain label keys.
package dltree

// Tree is a red-black tree for key-value pairs where key is domain label.
type Tree struct {
	root *node
}

// Pair is a key-value pair representing tree node content.
type Pair struct {
	Key   string
	Value interface{}
}

// RawPair is a key-value pair representing tree node content. Exposes binary representation of domain label.
type RawPair struct {
	Key   DomainLabel
	Value interface{}
}

// NewTree creates empty tree.
func NewTree() *Tree {
	return new(Tree)
}

// Insert puts given key-value pair to the tree and returns pointer to new root.
func (t *Tree) Insert(key string, value interface{}) *Tree {
	var (
		n *node
	)

	if t != nil {
		n = t.root
	}

	dl, _ := MakeDomainLabel(key)
	return &Tree{root: n.insert(dl, value)}
}

// RawInsert puts given key-value pair to the tree and returns pointer to new root. Expects bindary domain label on input.
func (t *Tree) RawInsert(key []byte, value interface{}) *Tree {
	var (
		n *node
	)

	if t != nil {
		n = t.root
	}

	return &Tree{root: n.insert(DomainLabel(key), value)}
}

// InplaceInsert inserts or replaces given key-value pair in the tree. The method inserts data directly to current tree so make sure you have exclusive access to it.
func (t *Tree) InplaceInsert(key string, value interface{}) {
	dl, _ := MakeDomainLabel(key)
	t.root = t.root.inplaceInsert(dl, value)
}

// RawInplaceInsert inserts or replaces given key-value pair in the tree. The method inserts data directly to current tree so make sure you have exclusive access to it. Expects bindary domain label on input.
func (t *Tree) RawInplaceInsert(key []byte, value interface{}) {
	t.root = t.root.inplaceInsert(DomainLabel(key), value)
}

// Get returns value by given key.
func (t *Tree) Get(key string) (interface{}, bool) {
	if t == nil {
		return nil, false
	}

	dl, _ := MakeDomainLabel(key)
	return t.root.get(dl)
}

// RawGet returns value by given key. Expects bindary domain label on input.
func (t *Tree) RawGet(key []byte) (interface{}, bool) {
	if t == nil {
		return nil, false
	}

	return t.root.get(DomainLabel(key))
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

// RawEnumerate returns channel which is populated by key pair values in order of keys. Returns binary domain labels.
func (t *Tree) RawEnumerate() chan RawPair {
	ch := make(chan RawPair)

	go func() {
		defer close(ch)

		if t == nil {
			return
		}

		t.root.rawEnumerate(ch)
	}()

	return ch
}

// Delete removes node by given key. It returns copy of tree and true if node has been indeed deleted otherwise copy of tree and false.
func (t *Tree) Delete(key string) (*Tree, bool) {
	if t == nil {
		return nil, false
	}

	dl, _ := MakeDomainLabel(key)
	root, ok := t.root.del(dl)
	return &Tree{root: root}, ok
}

// RawDelete removes node by given key. It returns copy of tree and true if node has been indeed deleted otherwise copy of tree and false. Expects bindary domain label on input.
func (t *Tree) RawDelete(key []byte) (*Tree, bool) {
	if t == nil {
		return nil, false
	}

	root, ok := t.root.del(DomainLabel(key))
	return &Tree{root: root}, ok
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
