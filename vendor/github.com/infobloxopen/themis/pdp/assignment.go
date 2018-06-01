package pdp

// AttributeAssignmentExpression represents assignment of arbitrary expression
// result to an attribute.
type AttributeAssignmentExpression struct {
	a Attribute
	e Expression
}

// MakeAttributeAssignmentExpression creates attribute assignment expression.
func MakeAttributeAssignmentExpression(a Attribute, e Expression) AttributeAssignmentExpression {
	return AttributeAssignmentExpression{
		a: a,
		e: e,
	}
}

// Serialize evaluates assignment expression and returns string representation
// of resulting attribute name, type and value or error if the evaluaction
// can't be done.
func (a AttributeAssignmentExpression) Serialize(ctx *Context) (string, string, string, error) {
	ID := a.a.id
	k := a.a.GetType().GetKey()

	v, err := a.e.Calculate(ctx)
	if err != nil {
		return ID, k, "", bindErrorf(err, "assignment to %q", ID)
	}

	t := v.GetResultType()
	if a.a.GetType() != t {
		return ID, k, "", bindErrorf(newAssignmentTypeMismatch(a.a, t), "assignment to %q", ID)
	}

	s, err := v.Serialize()
	if err != nil {
		return ID, k, "", bindErrorf(err, "assignment to %q", ID)
	}

	return ID, k, s, nil
}
