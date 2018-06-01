package pdp

// AttributeDesignator represents an expression which result is corresponding
// attribute value from request context.
type AttributeDesignator struct {
	a Attribute
}

// MakeAttributeDesignator creates designator expression instance for given
// attribute.
func MakeAttributeDesignator(a Attribute) AttributeDesignator {
	return AttributeDesignator{a}
}

// GetID returns ID of wrapped attribute.
func (d AttributeDesignator) GetID() string {
	return d.a.id
}

// GetResultType returns type of wrapped attribute (implements Expression
// interface).
func (d AttributeDesignator) GetResultType() Type {
	return d.a.t
}

// Calculate implements Expression interface and returns calculated value
func (d AttributeDesignator) Calculate(ctx *Context) (AttributeValue, error) {
	return ctx.getAttribute(d.a)
}
