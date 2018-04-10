package pdp

import "fmt"

type functionSetOfStringsContains struct {
	set   Expression
	value Expression
}

func makeFunctionSetOfStringsContains(set, value Expression) Expression {
	return functionSetOfStringsContains{
		set:   set,
		value: value}
}

func makeFunctionSetOfStringsContainsAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"contains\" for Set Of Strings needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionSetOfStringsContains(args[0], args[1])
}

func (f functionSetOfStringsContains) GetResultType() int {
	return TypeBoolean
}

func (f functionSetOfStringsContains) describe() string {
	return "contains"
}

// Calculate implements Expression interface and returns calculated value
func (f functionSetOfStringsContains) Calculate(ctx *Context) (AttributeValue, error) {
	set, err := ctx.calculateSetOfStringsExpression(f.set)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	s, err := ctx.calculateStringExpression(f.value)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	_, ok := set.Get(s)
	return MakeBooleanValue(ok), nil
}

func functionSetOfStringsContainsValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeSetOfStrings || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionSetOfStringsContainsAlt
}
