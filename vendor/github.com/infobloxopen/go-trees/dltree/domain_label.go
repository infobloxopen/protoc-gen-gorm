package dltree

import (
	"bytes"
	"math"
	"strconv"
)

// A DomainLabel represents content of a label.
type DomainLabel []byte

// GetFirstLabelSize returns size in bytes needed to store first label of given domain name as a DomainLabel. Additionally the function returns position right after the label in given string (or length of the string if the first label also is the last).
func GetFirstLabelSize(s string) (int, int) {
	size := 0
	escaped := 0
	var code [3]byte

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped <= 0 {
			switch c {
			case '.':
				return size, i

			case '\\':
				escaped = 1

			default:
				size++
			}
		} else if escaped == 1 {
			if c >= '0' && c <= '9' {
				code[0] = c
				escaped++
			} else {
				size++
				escaped = 0
			}
		} else if escaped > 1 && escaped < 4 {
			if c >= '0' && c <= '9' {
				code[escaped-1] = c
				escaped++
			} else {
				size += escaped
				escaped = 0

				switch c {
				case '.':
					return size, i

				case '\\':
					escaped = 1

				default:
					size++
				}
			}
		} else {
			if n, _ := strconv.Atoi(string(code[:])); n >= 0 && n <= math.MaxUint8 {
				size++
			} else {
				size += escaped
			}

			escaped = 0

			switch c {
			case '.':
				return size, i

			case '\\':
				escaped = 1

			default:
				size++
			}
		}
	}

	if escaped > 0 && escaped < 4 {
		size += escaped
	} else if escaped >= 4 {
		if n, _ := strconv.Atoi(string(code[:])); n >= 0 && n <= math.MaxUint8 {
			size++
		} else {
			size += escaped
		}
	}

	return size, len(s)
}

// MakeDomainLabel returns first domain label found in given string as DomainLabel and position in the string right after the label.
func MakeDomainLabel(s string) (DomainLabel, int) {
	size, end := GetFirstLabelSize(s)
	out := make(DomainLabel, size)

	escaped := 0
	var code [3]byte
	p := 0

	for i := 0; i < len(s); i++ {
		c := s[i]
		if escaped <= 0 {
			switch c {
			case '.':
				return out, end

			case '\\':
				escaped = 1

			default:
				if c >= 'A' && c <= 'Z' {
					c += 0x20
				}
				out[p] = c
				p++
			}
		} else if escaped == 1 {
			if c >= '0' && c <= '9' {
				code[0] = c
				escaped++
			} else {
				if c >= 'A' && c <= 'Z' {
					c += 0x20
				}
				out[p] = c
				p++

				escaped = 0
			}
		} else if escaped > 1 && escaped < 4 {
			if c >= '0' && c <= '9' {
				code[escaped-1] = c
				escaped++
			} else {
				out[p] = '\\'
				p++

				for j := 0; j < escaped-1; j++ {
					out[p] = code[j]
					p++
				}

				escaped = 0

				switch c {
				case '.':
					return out, end

				case '\\':
					escaped = 1

				default:
					if c >= 'A' && c <= 'Z' {
						c += 0x20
					}
					out[p] = c
					p++
				}
			}
		} else {
			if n, _ := strconv.Atoi(string(code[:])); n >= 0 && n <= math.MaxUint8 {
				out[p] = byte(n)
				if out[p] >= 'A' && out[p] <= 'Z' {
					out[p] += 0x20
				}
				p++
			} else {
				out[p] = '\\'
				p++

				for _, b := range code {
					out[p] = b
					p++
				}
			}

			escaped = 0

			switch c {
			case '.':
				return out, end

			case '\\':
				escaped = 1

			default:
				if c >= 'A' && c <= 'Z' {
					c += 0x20
				}
				out[p] = c
				p++
			}
		}
	}

	if escaped > 0 && escaped < 4 {
		out[p] = '\\'
		p++

		for i := 0; i < escaped-1; i++ {
			out[p] = code[i]
			p++
		}
	} else if escaped >= 4 {
		if n, _ := strconv.Atoi(string(code[:])); n >= 0 && n <= math.MaxUint8 {
			out[p] = byte(n)
			if out[p] >= 'A' && out[p] <= 'Z' {
				out[p] += 0x20
			}
			p++
		} else {
			out[p] = '\\'
			p++

			for _, b := range code {
				out[p] = b
				p++
			}
		}
	}

	return out, end
}

// String returns domain label in human readable format.
func (l DomainLabel) String() string {
	size := 0
	for _, c := range l {
		size++
		if c < ' ' || c > '~' {
			size += 3
		} else if c == '.' || c == '\\' {
			size++
		}
	}

	out := make([]byte, size)
	i := 0
	for _, c := range l {
		if c < ' ' || c > '~' {
			out[i] = '\\'

			if c < 100 {
				i++
				out[i] = '0'
				if c < 10 {
					i++
					out[i] = '0'
				}
			}

			for _, n := range strconv.Itoa(int(c)) {
				i++
				out[i] = byte(n)
			}
		} else {
			if c == '.' || c == '\\' {
				out[i] = '\\'
				i++
			}

			out[i] = c
		}

		i++
	}

	return string(out)
}

func compare(a, b DomainLabel) int {
	d := len(a) - len(b)
	if d != 0 {
		return d
	}

	return bytes.Compare(a, b)
}
