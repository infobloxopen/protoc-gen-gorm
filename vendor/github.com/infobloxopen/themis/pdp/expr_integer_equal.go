package pdp

import "fmt"

type functionIntegerEqual struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerEqual(first, second Expression) Expression {
	return functionIntegerEqual{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerEqual(args[0], args[1])
}

func (f functionIntegerEqual) GetResultType() int {
	return TypeBoolean
}

func (f functionIntegerEqual) describe() string {
	return "equal"
}

func (f functionIntegerEqual) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeBooleanValue(first == second), nil
}

func functionIntegerEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerEqualAlt
}
