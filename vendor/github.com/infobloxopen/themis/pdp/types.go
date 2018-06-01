package pdp

import (
	"sort"
	"strconv"
	"strings"
)

// Type* values represent all built-in data types PDP can work with.
var (
	// TypeUndefined stands for type of undefined value. The value usually
	// means that evaluation can't be done.
	TypeUndefined = newBuiltinType("Undefined")
	// TypeBoolean is boolean data type.
	TypeBoolean = newBuiltinType("Boolean")
	// TypeString is string data type.
	TypeString = newBuiltinType("String")
	// TypeInteger is integer data type.
	TypeInteger = newBuiltinType("Integer")
	// TypeFloat is float data type.
	TypeFloat = newBuiltinType("Float")
	// TypeAddress is IPv4 or IPv6 address data type.
	TypeAddress = newBuiltinType("Address")
	// TypeNetwork is IPv4 or IPv6 network data type.
	TypeNetwork = newBuiltinType("Network")
	// TypeDomain is domain name data type.
	TypeDomain = newBuiltinType("Domain")
	// TypeSetOfStrings is set of strings data type (internally stores order
	// in which it was created).
	TypeSetOfStrings = newBuiltinType("Set of Strings")
	// TypeSetOfNetworks is set of networks data type (unordered).
	TypeSetOfNetworks = newBuiltinType("Set of Networks")
	// TypeSetOfDomains is set of domains data type (unordered).
	TypeSetOfDomains = newBuiltinType("Set of Domains")
	// TypeListOfStrings is list of strings data type.
	TypeListOfStrings = newBuiltinType("List of Strings")

	// BuiltinTypeIDs maps type keys to Type* constants.
	BuiltinTypes = make(map[string]Type)
)

// Type is generic data type.
type Type interface {
	// String returns human readable type name.
	String() string
	// GetKey returns case insensitive (always lowercase) type key.
	GetKey() string
	// Match checks if type matches to other type. Built-in types match
	// iff they are equal.
	Match(t Type) bool
}

type builtinType struct {
	n string
	k string
}

func newBuiltinType(s string) Type {
	t := &builtinType{
		n: s,
		k: strings.ToLower(s),
	}

	BuiltinTypes[t.GetKey()] = t

	return t
}

func (t *builtinType) String() string {
	return t.n
}

func (t *builtinType) GetKey() string {
	return t.k
}

func (t *builtinType) Match(ot Type) bool {
	return t == ot
}

// FlagsType instance represents cutom flags type.
type FlagsType struct {
	n string
	k string
	f map[string]int
	b []string
	c int
}

// NewFlagsType function creates new custom type with given name. A value of
// the type can take any combination of listed flags (including empty set).
// It supports up to 64 flags and flag names should be unique for the type.
func NewFlagsType(name string, flags ...string) (Type, error) {
	key := strings.ToLower(name)
	if _, ok := BuiltinTypes[key]; ok {
		return nil, newDuplicatesBuiltinTypeError(name)
	}

	if len(flags) <= 0 {
		return nil, newNoFlagsDefinedError(name, len(flags))
	}

	if len(flags) > 64 {
		return nil, newTooManyFlagsDefinedError(name, len(flags))
	}

	c := 8
	if len(flags) > 8 {
		c = 16
	}
	if len(flags) > 16 {
		c = 32
	}
	if len(flags) > 32 {
		c = 64
	}

	f := make(map[string]int, len(flags))
	for i, s := range flags {
		n := strings.ToLower(s)
		flags[i] = n

		if j, ok := f[n]; ok {
			return nil, newDuplicateFlagName(name, s, i, j)
		}
		f[n] = i
	}

	return &FlagsType{
		n: name,
		k: strings.ToLower(name),
		f: f,
		b: flags,
		c: c,
	}, nil
}

// String method returns human readable type name.
func (t *FlagsType) String() string {
	return t.n
}

// GetKey method returns case insensitive (always lowercase) type key.
func (t *FlagsType) GetKey() string {
	return t.k
}

// Match checks equivalence of different flags types. Flags types match iff
// they are defined for the same number of flags.
func (t *FlagsType) Match(ot Type) bool {
	fot, ok := ot.(*FlagsType)
	if !ok {
		return false
	}

	if t == fot {
		return true
	}

	return len(t.b) == len(fot.b)
}

// Capacity gets number of bits required to represent any flags combination.
func (t *FlagsType) Capacity() int {
	return t.c
}

// GetFlagBit method returns bit number for given flag name. If there is no flag
// with the name it returns -1.
func (t *FlagsType) GetFlagBit(f string) int {
	if n, ok := t.f[f]; ok {
		return n
	}

	return -1
}

// Signature is an ordered sequence of types.
type Signature []Type

// MakeSignature function creates signature from given types.
func MakeSignature(t ...Type) Signature {
	return t
}

// String method returns a string containing list of types separated by slash.
func (s Signature) String() string {
	if len(s) > 0 {
		seq := make([]string, len(s))
		for i, t := range s {
			seq[i] = strconv.Quote(t.String())
		}

		return strings.Join(seq, "/")
	}

	return "empty"
}

// TypeSet represent an unordered set of types.
type TypeSet map[Type]struct{}

func makeTypeSet(t ...Type) TypeSet {
	s := make(TypeSet, len(t))
	for _, t := range t {
		s[t] = struct{}{}
	}
	return s
}

// Contains method checks whether the set contains a type.
func (s TypeSet) Contains(t Type) bool {
	_, ok := s[t]
	return ok
}

// String method returns a string containing type names separated by comma.
func (s TypeSet) String() string {
	if len(s) > 0 {
		seq := make([]string, len(s))
		i := 0
		for t := range s {
			seq[i] = strconv.Quote(t.String())
			i++
		}

		sort.Strings(seq)
		return strings.Join(seq, ", ")
	}

	return "empty"
}
