package domain

import (
	"fmt"
	"math"
)

const (
	escRegular = iota
	escFirstChar
	escSecondDigit
	escThirdDigit
)

// MakeLabel makes uppercase domain label from given human-readable representation. Ignores ending dot.
func MakeLabel(s string) (string, error) {
	var label [MaxLabel + 1]byte

	n, err := getLabel(s, label[:])
	if err != nil {
		return "", err
	}

	return string(label[1:n]), nil
}

// MakeHumanReadableLabel makes human-readable label by escaping according RFC-4343.
func MakeHumanReadableLabel(s string) string {
	var label [4 * MaxLabel]byte

	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' || c == '\\' {
			if j >= len(label)-1 {
				return string(label[:j])
			}

			label[j] = '\\'
			label[j+1] = c
			j += 2

		} else if c < '!' || c > '~' {
			if j >= len(label)-3 {
				return string(label[:j])
			}

			label[j] = '\\'

			r := c % 10
			label[j+3] = r + '0'

			c /= 10
			r = c % 10
			label[j+2] = r + '0'

			c /= 10
			label[j+1] = c + '0'

			j += 4
		} else if c >= 'A' && c <= 'Z' {
			label[j] = c | 0x20
			j++
		} else {
			if j >= len(label) {
				return string(label[:j])
			}

			label[j] = c
			j++
		}
	}

	return string(label[:j])
}

func markLabels(s string, offs []int) (int, error) {
	n := 0
	start := 0
	esc := escRegular
	var code int
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch esc {
		case escRegular:
			switch c {
			case '.':
				if start >= i {
					return 0, ErrEmptyLabel
				}

				if n >= len(offs) {
					return 0, ErrTooManyLabels
				}

				offs[n] = start
				n++

				start = i + 1

			case '\\':
				esc = escFirstChar
			}

		case escFirstChar:
			if c < '0' || c > '9' {
				esc = escRegular
			} else {
				code = int(c-'0') * 100
				if code > math.MaxUint8 {
					return 0, ErrInvalidEscape
				}

				esc = escSecondDigit
			}

		case escSecondDigit:
			if c < '0' || c > '9' {
				return 0, ErrInvalidEscape
			}

			code += int(c-'0') * 10
			if code > math.MaxUint8 {
				return 0, ErrInvalidEscape
			}

			esc = escThirdDigit

		case escThirdDigit:
			if c < '0' || c > '9' {
				return 0, ErrInvalidEscape
			}

			code += int(c - '0')
			if code > math.MaxUint8 {
				return 0, ErrInvalidEscape
			}

			esc = escRegular
		}
	}

	if esc != escRegular {
		return 0, ErrInvalidEscape
	}

	if start < len(s) {
		if n >= len(offs) {
			return 0, ErrTooManyLabels
		}

		offs[n] = start
		n++
	}

	return n, nil
}

func getLabel(s string, out []byte) (int, error) {
	j := 1
	esc := escRegular
	var code int

Loop:
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch esc {
		case escRegular:
			switch c {
			default:
				if c >= 'a' && c <= 'z' {
					c &= 0xdf
				}

				if j >= len(out) {
					return 0, ErrLabelTooLong
				}

				out[j] = c
				j++

			case '.':
				if i < len(s)-1 {
					panic(fmt.Errorf("unescaped dot at %d index before last character %d", i, len(s)-1))
				}

				break Loop

			case '\\':
				esc = escFirstChar
			}

		case escFirstChar:
			if c < '0' || c > '9' {
				if c >= 'a' && c <= 'z' {
					c &= 0xdf
				}

				if j >= len(out) {
					return 0, ErrLabelTooLong
				}

				out[j] = c
				j++

				esc = escRegular
			} else {
				code = int(c-'0') * 100
				if code > math.MaxUint8 {
					return 0, ErrInvalidEscape
				}

				esc = escSecondDigit
			}

		case escSecondDigit:
			if c < '0' || c > '9' {
				return 0, ErrInvalidEscape
			}

			code += int(c-'0') * 10
			if code > math.MaxUint8 {
				return 0, ErrInvalidEscape
			}

			esc = escThirdDigit

		case escThirdDigit:
			if c < '0' || c > '9' {
				return 0, ErrInvalidEscape
			}

			code += int(c - '0')
			if code > math.MaxUint8 {
				return 0, ErrInvalidEscape
			}

			c = byte(code)
			if c >= 'a' && c <= 'z' {
				c &= 0xdf
			}

			if j >= len(out) {
				return 0, ErrLabelTooLong
			}

			out[j] = c
			j++

			esc = escRegular
		}
	}

	if esc != escRegular {
		return 0, ErrInvalidEscape
	}

	out[0] = byte(j - 1)

	return j, nil
}
