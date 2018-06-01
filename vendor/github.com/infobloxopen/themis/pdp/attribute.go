package pdp

import "fmt"

// Attribute represents attribute definition which binds attribute name
// and type.
type Attribute struct {
	id string
	t  Type
}

// MakeAttribute creates new attribute instance. It requires attribute name
// as "ID" argument and type as "t" argument.
func MakeAttribute(ID string, t Type) Attribute {
	return Attribute{id: ID, t: t}
}

// GetType returns attribute type.
func (a Attribute) GetType() Type {
	return a.t
}

func (a Attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, a.t)
}
