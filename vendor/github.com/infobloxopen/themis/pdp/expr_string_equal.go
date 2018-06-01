package pdp

import "fmt"

type functionStringEqual struct {
	first  Expression
	second Expression
}

func makeFunctionStringEqual(first, second Expression) Expression {
	return functionStringEqual{
		first:  first,
		second: second}
}

func makeFunctionStringEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for String needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionStringEqual(args[0], args[1])
}

func (f functionStringEqual) GetResultType() Type {
	return TypeBoolean
}

func (f functionStringEqual) describe() string {
	return "equal"
}

// Calculate implements Expression interface and returns calculated value
func (f functionStringEqual) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateStringExpression(f.first)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateStringExpression(f.second)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeBooleanValue(first == second), nil
}

func functionStringEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeString || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionStringEqualAlt
}
