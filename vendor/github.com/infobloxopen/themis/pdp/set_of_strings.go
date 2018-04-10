package pdp

import (
	"sort"

	"github.com/infobloxopen/go-trees/strtree"
)

func sortSetOfStrings(v *strtree.Tree) []string {
	pairs := newPairList(v)
	sort.Sort(pairs)

	list := make([]string, len(pairs))
	for i, pair := range pairs {
		list[i] = pair.value
	}

	return list
}

type pair struct {
	value string
	order int
}

type pairList []pair

func newPairList(v *strtree.Tree) pairList {
	pairs := make(pairList, 0)
	for p := range v.Enumerate() {
		pairs = append(pairs, pair{p.Key, p.Value.(int)})
	}

	return pairs
}

func (p pairList) Len() int {
	return len(p)
}

func (p pairList) Less(i, j int) bool {
	return p[i].order < p[j].order
}

func (p pairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
