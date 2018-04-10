package domaintree

import (
	"errors"

	"github.com/infobloxopen/go-trees/dltree"
)

var (
	// ErrCompressedDN is the error returned by WireGet when domain label exceeds 63 bytes.
	ErrCompressedDN = errors.New("can't handle compressed domain name")
	// ErrLabelTooLong is the error returned by WireGet when last domain label length doesn't
	// fit whole domain name length.
	ErrLabelTooLong = errors.New("label too long")
	// ErrEmptyLabel means that label of zero length met in the middle of domain name.
	ErrEmptyLabel = errors.New("empty label")
	// ErrNameTooLong is the error returned when overall domain name length exeeds 256 bytes.
	ErrNameTooLong = errors.New("domain name too long")
)

func split(s string) []dltree.DomainLabel {
	dn := make([]dltree.DomainLabel, getLabelsCount(s))
	if len(dn) > 0 {
		end := len(dn) - 1
		start := 0
		for i := range dn {
			label, p := dltree.MakeDomainLabel(s[start:])
			start += p + 1
			dn[end-i] = label
		}
	}

	return dn
}

func getLabelsCount(s string) int {
	labels := 0
	start := 0
	for {
		size, p := dltree.GetFirstLabelSize(s[start:])
		start += p + 1
		if start >= len(s) {
			if size > 0 {
				labels++
			}

			break
		}

		labels++
	}

	return labels
}

// WireDomainNameLower is a type to store domain name in "wire" format as described in RFC-1035 section "3.1. Name space definitions" with all lowercase ASCII letters.
type WireDomainNameLower []byte

// MakeWireDomainNameLower creates lowercase "wire" representation of given domain name.
func MakeWireDomainNameLower(s string) (WireDomainNameLower, error) {
	out := WireDomainNameLower{}
	start := 0
	for {
		label, p := dltree.MakeDomainLabel(s[start:])
		if len(label) > 63 {
			return nil, ErrLabelTooLong
		}

		start += p + 1
		if start >= len(s) {
			if len(label) > 0 {
				out = append(out, byte(len(label)))
				out = append(out, label...)
				if len(out) > 255 {
					return nil, ErrNameTooLong
				}
			}

			break
		}

		if len(label) <= 0 {
			return nil, ErrEmptyLabel
		}

		out = append(out, byte(len(label)))
		out = append(out, label...)
		if len(out) > 255 {
			return nil, ErrNameTooLong
		}
	}

	return append(out, 0), nil
}

// ToLowerWireDomainName converts "wire" domain name to lowercase.
func ToLowerWireDomainName(d []byte) (WireDomainNameLower, error) {
	if len(d) > 256 {
		return nil, ErrNameTooLong
	}

	out := make(WireDomainNameLower, len(d))
	ll := 0
	for i, c := range d {
		if ll > 0 {
			ll--

			if c >= 'A' && c <= 'Z' {
				c += 0x20
			}
		} else {
			ll = int(c)
			if ll <= 0 && i != len(d)-1 {
				return nil, ErrEmptyLabel
			}

			if ll > 63 {
				return nil, ErrCompressedDN
			}
		}

		out[i] = c
	}

	if out[len(out)-1] != 0 {
		return nil, ErrLabelTooLong
	}

	return out, nil
}

// String returns domain name in human readable format.
func (d WireDomainNameLower) String() string {
	out := ""
	start := 0
	for start < len(d) {
		ll := int(d[start])

		start++
		if start >= len(d) {
			if ll > 0 {
				out += "."
			}

			return out
		}

		if ll > 0 {
			label := dltree.DomainLabel(d[start : start+ll]).String()
			if len(out) > 0 {
				out += "." + label
			} else {
				out = label
			}

			start += ll
		}
	}

	return out
}

func wireSplitCallback(dn WireDomainNameLower, f func(label []byte) bool) error {
	if len(dn) > 256 {
		return ErrNameTooLong
	}

	if len(dn) > 0 {
		var lPos [256]int
		labels := 0
		idx := 0
		max := 0
		for {
			ll := int(dn[idx])
			if ll <= 0 {
				if idx != len(dn)-1 {
					return ErrEmptyLabel
				}

				break
			}

			if ll > 63 {
				return ErrCompressedDN
			}

			if idx+ll+1 > len(dn) {
				return ErrLabelTooLong
			}

			if ll > max {
				max = ll
			}

			lPos[labels] = idx
			labels++
			idx += ll + 1
		}

		for labels > 0 {
			labels--
			idx := lPos[labels]
			ll := int(dn[idx])
			start := idx + 1
			end := start + ll
			if !f(dn[start:end]) {
				break
			}
		}
	}

	return nil
}
