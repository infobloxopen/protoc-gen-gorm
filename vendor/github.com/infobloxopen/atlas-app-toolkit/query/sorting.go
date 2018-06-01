package query

import (
	"fmt"
	"strings"
)

// IsAsc returns true if sort criteria has ascending sort order, otherwise false.
func (c SortCriteria) IsAsc() bool {
	return c.Order == SortCriteria_ASC
}

// IsDesc returns true if sort criteria has descending sort order, otherwise false.
func (c SortCriteria) IsDesc() bool {
	return c.Order == SortCriteria_DESC
}

// GoString implements fmt.GoStringer interface
// return string representation of a sort criteria in next form:
// "<tag_name> (ASC|DESC)".
func (c SortCriteria) GoString() string {
	return fmt.Sprintf("%s %s", c.Tag, c.Order)
}

// ParseSorting parses raw string that represent sort criteria into a Sorting
// data structure.
// Provided string is supposed to be in accordance with the sorting collection
// operator from REST API Syntax.
// See: https://github.com/infobloxopen/atlas-app-toolkit#sorting
func ParseSorting(s string) (*Sorting, error) {
	var sorting Sorting

	for _, craw := range strings.Split(s, ",") {
		v := strings.Fields(craw)

		var c SortCriteria
		switch len(v) {
		case 1:
			c.Tag, c.Order = v[0], SortCriteria_ASC
		case 2:
			if o, ok := SortCriteria_Order_value[strings.ToUpper(v[1])]; !ok {
				return nil, fmt.Errorf("invalid sort order - %q in %q", v[1], craw)
			} else {
				c.Tag, c.Order = v[0], SortCriteria_Order(o)
			}
		default:
			return nil, fmt.Errorf("invalid sort criteria: %s", craw)
		}

		sorting.Criterias = append(sorting.Criterias, &c)
	}

	return &sorting, nil
}

// GoString implements fmt.GoStringer interface
// Returns string representation of sorting in next form:
// "<name> (ASC|DESC) [, <tag_name> (ASC|DESC)]"
func (s Sorting) GoString() string {
	var l []string

	for _, c := range s.GetCriterias() {
		l = append(l, c.GoString())
	}

	return strings.Join(l, ", ")
}
