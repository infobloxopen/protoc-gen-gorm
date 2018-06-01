// Package domain provide functions to parse and handle domain names and labels.
package domain

import "errors"

var (
	// ErrTooManyLabels is returned for domain name with more than 127 labels.
	ErrTooManyLabels = errors.New("too many labels")
	// ErrEmptyLabel indicates that domain name contains empty label.
	ErrEmptyLabel = errors.New("empty label")
	// ErrLabelTooLong is returned when one of domain labels has more than 63 characters.
	ErrLabelTooLong = errors.New("label too long")
	// ErrNameTooLong is returned when domain name has more than 255 characters.
	ErrNameTooLong = errors.New("name too long")
	// ErrInvalidEscape is returned for invalid escape sequence.
	ErrInvalidEscape = errors.New("invalid escape sequence")
)

// Name is a structure which represents domain name.
type Name struct {
	h string
	c string
}

const (
	// MaxName is maximum number of bytes for whole domain name.
	MaxName = 255
	// MaxLabels is maximum number of labels domain name can consist of.
	MaxLabels = MaxName / 2
	// MaxLabel is maximim number of bytes for single label.
	MaxLabel = 63
)

// MakeNameFromString creates a Name from human-readable domain name string.
func MakeNameFromString(s string) (Name, error) {
	out := Name{h: s}

	if len(s) < 1 {
		return out, nil
	}

	if len(s) == 1 && s[0] == '.' {
		return out, nil
	}

	var offs [MaxLabels]int

	n, err := markLabels(s, offs[:])
	if err != nil {
		return out, err
	}

	var (
		label [MaxLabel + 1]byte
		name  [MaxName]byte
	)

	j := 0
	end := len(s)
	for i := n - 1; i >= 0; i-- {
		start := offs[i]

		n, err := getLabel(s[start:end], label[:])
		if err != nil {
			return out, err
		}

		if copied := copy(name[j:], label[:n]); copied < n {
			return out, ErrNameTooLong
		}

		j += n

		end = start
	}

	out.c = string(name[:j])

	return out, nil
}

// String method returns domain name in human-readable format.
func (n Name) String() string {
	return n.h
}

// GetLabel returns label starting from given offset and offset of the next label. The method returns zero offset for the last label and -1 in case of error.
func (n Name) GetLabel(off int) (string, int) {
	if off == 0 && len(n.c) == 0 {
		return "", 0
	}

	if off < 0 || off >= len(n.c) {
		return "", -1
	}

	size := int(n.c[off])
	if size < 1 || size > 63 {
		return "", -1
	}

	start := off + 1
	end := start + size
	if end >= len(n.c) {
		return n.c[start:], 0
	}

	return n.c[start:end], end
}

// GetLabels iterate through name labels in reversed order.
func (n Name) GetLabels(f func(string) error) error {
	off := 0
	for off < len(n.c) {
		off++
		next := off + int(n.c[off-1])

		if err := f(n.c[off:next]); err != nil {
			return err
		}

		off = next
	}

	return nil
}
