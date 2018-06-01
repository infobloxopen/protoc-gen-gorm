package query

import (
	"fmt"
	"strconv"
)

const (
	// Default pagination limit
	DefaultLimit = 1000

	lastOffset = int32(1 << 30)
)

// Pagination parses string representation of pagination limit, offset.
// Returns error if limit or offset has invalid syntax or out of range.
func ParsePagination(limit, offset, ptoken string) (*Pagination, error) {
	p := new(Pagination)

	if limit != "" {
		if u, err := strconv.ParseInt(limit, 10, 32); err != nil {
			return nil, fmt.Errorf("pagination: limit - %s", err.(*strconv.NumError).Err)
		} else if u < 0 {
			return nil, fmt.Errorf("pagination: limit - negative value")
		} else {
			p.Limit = int32(u)
		}
	}

	if offset == "null" {
		p.Offset = 0
	} else if offset != "" {
		if u, err := strconv.ParseInt(offset, 10, 32); err != nil {
			return nil, fmt.Errorf("pagination: offset - %s", err.(*strconv.NumError).Err)
		} else if u < 0 {
			return nil, fmt.Errorf("pagination: offset - negative value")
		} else {
			p.Offset = int32(u)
		}
	}

	if ptoken != "" {
		p.PageToken = ptoken
	}

	return p, nil
}

// FirstPage returns true if requested first page
func (p *Pagination) FirstPage() bool {
	if p.GetPageToken() == "null" || p.GetOffset() == 0 {
		return true
	}
	return false
}

// DefaultLimit returns DefaultLimit if limit was not specified otherwise
// returns either requested or specified one.
func (p *Pagination) DefaultLimit(dl ...int32) int32 {
	if l := p.GetLimit(); l != 0 {
		return l
	}
	if len(dl) > 0 && dl[0] > 0 {
		return dl[0]
	}
	return DefaultLimit
}

// SetLastToken sets page info to indicate no more pages are available
func (p *PageInfo) SetLastToken() {
	p.PageToken = "null"
}

// SetLastOffset sets page info to indicate no more pages are available
func (p *PageInfo) SetLastOffset() {
	p.Offset = lastOffset
}

// NoMore reports whether page info indicates no more pages are available
func (p *PageInfo) NoMore() bool {
	if p.GetOffset() == lastOffset || p.GetPageToken() == "null" {
		return true
	}
	return false
}
