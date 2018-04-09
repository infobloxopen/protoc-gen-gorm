// Package domaintree implements radix tree data structure for domain names.
package domaintree

import "github.com/infobloxopen/go-trees/dltree"

// Node is a radix tree for domain names.
type Node struct {
	branches *dltree.Tree

	hasValue bool
	value    interface{}
}

// Pair represents a key-value pair returned by Enumerate method.
type Pair struct {
	Key   string
	Value interface{}
}

// Insert puts value using given domain as a key. The method returns new tree (old one remains unaffected). Input name converted to ASCII lowercase according to RFC-4343 (by mapping A-Z to a-z) to perform case-insensitive comparison when getting data from the tree.
func (n *Node) Insert(d string, v interface{}) *Node {
	if n == nil {
		n = &Node{}
	} else {
		n = &Node{
			branches: n.branches,
			hasValue: n.hasValue,
			value:    n.value}
	}
	r := n

	for _, label := range split(d) {
		item, ok := n.branches.RawGet(label)
		var next *Node
		if ok {
			next = item.(*Node)
			next = &Node{
				branches: next.branches,
				hasValue: next.hasValue,
				value:    next.value}
		} else {
			next = &Node{}
		}

		n.branches = n.branches.RawInsert(label, next)
		n = next
	}

	n.hasValue = true
	n.value = v

	return r
}

// InplaceInsert puts or replaces value using given domain as a key. The method inserts data directly to current tree so make sure you have exclusive access to it. Input name converted in the same way as for Insert.
func (n *Node) InplaceInsert(d string, v interface{}) {
	if n.branches == nil {
		n.branches = dltree.NewTree()
	}

	for _, label := range split(d) {
		item, ok := n.branches.RawGet(label)
		if ok {
			n = item.(*Node)
		} else {
			next := &Node{branches: dltree.NewTree()}
			n.branches.RawInplaceInsert(label, next)
			n = next
		}
	}

	n.hasValue = true
	n.value = v
}

// Enumerate returns key-value pairs in given tree sorted by key first by top level domain label second by second level and so on.
func (n *Node) Enumerate() chan Pair {
	ch := make(chan Pair)

	go func() {
		defer close(ch)
		n.enumerate("", ch)
	}()

	return ch
}

// Get gets value for domain which is equal to domain in the tree or is a subdomain of existing domain.
func (n *Node) Get(d string) (interface{}, bool) {
	if n == nil {
		return nil, false
	}

	for _, label := range split(d) {
		item, ok := n.branches.RawGet(label)
		if !ok {
			break
		}

		n = item.(*Node)
	}

	return n.value, n.hasValue
}

// WireGet gets value for domain which is equal to domain in the tree or is a subdomain of existing domain. The method accepts domain name in "wire" format described by RFC-1035 section "3.1. Name space definitions". Additionally it requires all ASCII letters (A-Z) to be converted to their lowercase counterparts (a-z). Returns error in case of compressed names (label length > 63 octets), malformed domain names (last label length too big) and too long domain names (more than 255 bytes).
func (n *Node) WireGet(d WireDomainNameLower) (interface{}, bool, error) {
	if n == nil {
		return nil, false, nil
	}

	err := wireSplitCallback(d, func(label []byte) bool {
		if item, ok := n.branches.RawGet(label); ok {
			n = item.(*Node)
			return true
		}

		return false
	})
	if err != nil {
		return nil, false, err
	}

	return n.value, n.hasValue, nil
}

// DeleteSubdomains removes current domain and all its subdomains if any. It returns new tree and flag if deletion indeed occurs.
func (n *Node) DeleteSubdomains(d string) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	labels := split(d)
	if len(labels) > 0 {
		return n.delSubdomains(split(d))
	}

	if n.hasValue || !n.branches.IsEmpty() {
		return &Node{}, true
	}

	return n, false
}

// Delete removes current domain only. It returns new tree and flag if deletion indeed occurs.
func (n *Node) Delete(d string) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	labels := split(d)
	if len(labels) > 0 {
		return n.del(split(d))
	}

	if n.hasValue || !n.branches.IsEmpty() {
		return &Node{}, true
	}

	return n, false
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

	for item := range n.branches.RawEnumerate() {
		sub := item.Key.String()
		if len(s) > 0 {
			sub += "." + s
		}
		node := item.Value.(*Node)

		node.enumerate(sub, ch)
	}
}

func (n *Node) delSubdomains(labels []dltree.DomainLabel) (*Node, bool) {
	label := labels[0]
	if len(labels) > 1 {
		item, ok := n.branches.RawGet(label)
		if !ok {
			return n, false
		}

		next := item.(*Node)
		next, ok = next.delSubdomains(labels[1:])
		if !ok {
			return n, false
		}

		if next.branches.IsEmpty() && !next.hasValue {
			branches, _ := n.branches.RawDelete(label)
			return &Node{
				branches: branches,
				hasValue: n.hasValue,
				value:    n.value}, true
		}

		return &Node{
			branches: n.branches.RawInsert(label, next),
			hasValue: n.hasValue,
			value:    n.value}, true
	}

	branches, ok := n.branches.RawDelete(label)
	if ok {
		return &Node{
			branches: branches,
			hasValue: n.hasValue,
			value:    n.value}, true
	}

	return n, false
}

func (n *Node) del(labels []dltree.DomainLabel) (*Node, bool) {
	label := labels[0]
	item, ok := n.branches.RawGet(label)
	if !ok {
		return n, false
	}
	next := item.(*Node)
	if len(labels) > 1 {
		next, ok = next.del(labels[1:])
		if !ok {
			return n, false
		}

		if next.branches.IsEmpty() && !next.hasValue {
			branches, _ := n.branches.RawDelete(label)
			return &Node{
				branches: branches,
				hasValue: n.hasValue,
				value:    n.value}, true
		}

		return &Node{
			branches: n.branches.RawInsert(label, next),
			hasValue: n.hasValue,
			value:    n.value}, true
	}

	if next.branches.IsEmpty() {
		branches, _ := n.branches.RawDelete(label)
		return &Node{
			branches: branches,
			hasValue: n.hasValue,
			value:    n.value}, true
	}

	return &Node{
		branches: n.branches.RawInsert(label, &Node{branches: next.branches}),
		hasValue: n.hasValue,
		value:    n.value}, true
}
