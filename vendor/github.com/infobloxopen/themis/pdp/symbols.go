package pdp

import "strings"

// Symbols wraps type and attribute symbol tables.
type Symbols struct {
	types map[string]Type
	attrs map[string]Attribute
	ro    bool
}

// MakeSymbols create symbol tables without any types and attributes.
func MakeSymbols() Symbols {
	return Symbols{
		types: make(map[string]Type),
		attrs: make(map[string]Attribute),
	}
}

// PutType stores given type in the symbol table.
func (s Symbols) PutType(t Type) error {
	if s.ro {
		return newReadOnlySymbolsChangeError()
	}

	if t == nil {
		return newNilTypeError()
	}

	if _, ok := t.(*builtinType); ok {
		return newBuiltinCustomTypeError(t)
	}

	k := t.GetKey()
	if prevT, ok := s.types[k]; ok {
		return newDuplicateCustomTypeError(t, prevT)
	}

	s.types[k] = t

	return nil
}

// GetType returns type by id. It can be built-in type or type stored in
// the symbol table.
func (s Symbols) GetType(ID string) Type {
	id := strings.ToLower(ID)
	if t, ok := BuiltinTypes[id]; ok {
		return t
	}

	if t, ok := s.types[id]; ok {
		return t
	}

	return nil
}

// PutAttribute stores given attribute in the symbol table.
func (s Symbols) PutAttribute(a Attribute) error {
	if s.ro {
		return newReadOnlySymbolsChangeError()
	}

	if a.t == nil {
		return newNoTypedAttributeError(a)
	}

	if a.t == TypeUndefined {
		return newUndefinedAttributeTypeError(a)
	}

	if s.GetType(a.t.GetKey()) == nil {
		return newUnknownAttributeTypeError(a)
	}

	if _, ok := s.attrs[a.id]; ok {
		return newDuplicateAttributeError(a)
	}

	s.attrs[a.id] = a

	return nil
}

// GetAttribute returns attribute by id.
func (s Symbols) GetAttribute(ID string) (Attribute, bool) {
	if a, ok := s.attrs[ID]; ok {
		return a, true
	}

	return Attribute{}, false
}

func (s Symbols) makeROCopy() Symbols {
	return Symbols{
		types: s.types,
		attrs: s.attrs,
		ro:    true,
	}
}
