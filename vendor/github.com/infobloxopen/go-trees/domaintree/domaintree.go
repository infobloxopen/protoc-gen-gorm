// Package domaintree implements radix tree data structure for domain names.
package domaintree

import (
	"errors"

	"github.com/infobloxopen/go-trees/dltree"
	"github.com/infobloxopen/go-trees/domain"
)

// Node is a radix tree for domain names.
type Node struct {
	branches *dltree.Tree

	hasValue bool
	value    interface{}
}

// Pair represents a key-value pair returned by Enumerate method.
type Pair struct {
	// Key is a human-readable representation of domain name.
	Key string
	// Value stores data related to the name.
	Value interface{}
}

var errStopIterations = errors.New("stop iterations")

// Insert puts value using given domain as a key. The method returns new tree (old one remains unaffected).
func (n *Node) Insert(d domain.Name, v interface{}) *Node {
	n = n.copy()
	r := n

	d.GetLabels(func(label string) error {
		item, ok := n.branches.RawGet(label)
		var next *Node
		if ok {
			next = item.(*Node).copy()
		} else {
			next = new(Node)
		}

		n.branches = n.branches.RawInsert(label, next)
		n = next

		return nil
	})

	n.hasValue = true
	n.value = v

	return r
}

// InplaceInsert puts or replaces value using given domain as a key. The method inserts data directly to current tree so make sure you have exclusive access to it.
func (n *Node) InplaceInsert(d domain.Name, v interface{}) {
	if n.branches == nil {
		n.branches = dltree.NewTree()
	}

	d.GetLabels(func(label string) error {
		item, ok := n.branches.RawGet(label)
		if ok {
			n = item.(*Node)
		} else {
			next := &Node{branches: dltree.NewTree()}
			n.branches.RawInplaceInsert(label, next)
			n = next
		}

		return nil
	})

	n.hasValue = true
	n.value = v
}

// Enumerate returns key-value pairs in given tree. It lists domains in the same order for the same tree.
func (n *Node) Enumerate() chan Pair {
	ch := make(chan Pair)

	go func() {
		defer close(ch)
		n.enumerate("", ch)
	}()

	return ch
}

// Get gets value for given domain which is equal to domain in the tree or is a subdomain of existing domain.
func (n *Node) Get(d domain.Name) (interface{}, bool) {
	if n == nil {
		return nil, false
	}

	d.GetLabels(func(label string) error {
		item, ok := n.branches.RawGet(label)
		if !ok {
			return errStopIterations
		}

		n = item.(*Node)
		return nil
	})

	return n.value, n.hasValue
}

// DeleteSubdomains removes current domain and all its subdomains if any. It returns new tree and flag if deletion indeed occurs.
func (n *Node) DeleteSubdomains(d domain.Name) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	var (
		labels [domain.MaxLabels]string
		nodes  [domain.MaxLabels]*Node
	)

	i := n.getBranch(d, labels[:], nodes[:])
	if i >= len(nodes) || !nodes[i].hasValue && n.branches.IsEmpty() {
		return n, false
	}

	i++
	if i >= len(nodes) {
		return new(Node), true
	}

	n = nodes[i].copy()
	n.branches, _ = n.branches.RawDelete(labels[i])
	i++

	return n.copyBranch(labels[i:], nodes[i:]), true
}

// Delete removes current domain only. It returns new tree and flag if deletion indeed occurs.
func (n *Node) Delete(d domain.Name) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	var (
		labels [domain.MaxLabels]string
		nodes  [domain.MaxLabels]*Node
	)

	i := n.getBranch(d, labels[:], nodes[:])
	if i >= len(nodes) || !nodes[i].hasValue {
		return n, false
	}

	n = nodes[i]
	i++

	branches := n.branches
	if i >= len(nodes) {
		if branches.IsEmpty() {
			return new(Node), true
		}

		return &Node{branches: branches}, true
	}

	n = nodes[i].copy()
	if branches.IsEmpty() {
		n.branches, _ = n.branches.RawDelete(labels[i])
	} else {
		n.branches = n.branches.RawInsert(labels[i], &Node{branches: branches})
	}
	i++

	return n.copyBranch(labels[i:], nodes[i:]), true
}

func (n *Node) copy() *Node {
	if n == nil {
		return new(Node)
	}

	return &Node{
		branches: n.branches,
		hasValue: n.hasValue,
		value:    n.value,
	}
}

func (n *Node) enumerate(s string, ch chan Pair) {
	if n == nil {
		return
	}

	if n.hasValue {
		ch <- Pair{
			Key:   s,
			Value: n.value}
	}

	for item := range n.branches.Enumerate() {
		sub := item.Key
		if len(s) > 0 {
			sub += "." + s
		}
		node := item.Value.(*Node)

		node.enumerate(sub, ch)
	}
}

func (n *Node) getBranch(d domain.Name, labels []string, nodes []*Node) int {
	i := len(labels) - 1
	nodes[i] = n

	if err := d.GetLabels(func(label string) error {
		labels[i] = label

		next, ok := n.branches.RawGet(label)
		if !ok {
			return errStopIterations
		}

		n = next.(*Node)

		i--
		nodes[i] = n
		return nil
	}); err != nil {
		return len(labels)
	}

	return i
}

func (n *Node) copyBranch(labels []string, nodes []*Node) *Node {
	for i, p := range nodes {
		p = p.copy()
		if !n.hasValue && n.branches.IsEmpty() {
			p.branches, _ = p.branches.RawDelete(labels[i])
		} else {
			p.branches = p.branches.RawInsert(labels[i], n)
		}

		n = p
	}

	return n
}
