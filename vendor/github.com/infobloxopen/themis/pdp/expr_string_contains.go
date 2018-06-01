package pdp

import (
	"fmt"
	"strings"
)

type functionStringContains struct {
	str    Expression
	substr Expression
}

func makeFunctionStringContains(str, substr Expression) Expression {
	return functionStringContains{
		str:    str,
		substr: substr}
}

func makeFunctionStringContainsAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"contains\" for String needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionStringContains(args[0], args[1])
}

func (f functionStringContains) GetResultType() Type {
	return TypeBoolean
}

func (f functionStringContains) describe() string {
	return "contains"
}

// Calculate implements Expression interface and returns calculated value
func (f functionStringContains) Calculate(ctx *Context) (AttributeValue, error) {
	str, err := ctx.calculateStringExpression(f.str)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "string argument"), f.describe())
	}

	substr, err := ctx.calculateStringExpression(f.substr)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "substring argument"), f.describe())
	}

	return MakeBooleanValue(strings.Contains(str, substr)), nil
}

func functionStringContainsValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeString || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionStringContainsAlt
}
