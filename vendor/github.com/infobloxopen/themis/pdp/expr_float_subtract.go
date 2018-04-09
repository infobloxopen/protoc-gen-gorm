package pdp

import "fmt"

type functionFloatSubtract struct {
	first  Expression
	second Expression
}

func makeFunctionFloatSubtract(first, second Expression) Expression {
	return functionFloatSubtract{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatSubtractAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"subtract\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatSubtract(args[0], args[1])
}

func (f functionFloatSubtract) GetResultType() int {
	return TypeFloat
}

func (f functionFloatSubtract) describe() string {
	return "subtract"
}

func (f functionFloatSubtract) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	res := first - second
	if err = floatErrorCheck(res); err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	return MakeFloatValue(res), nil
}

func functionFloatSubtractValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}
	return makeFunctionFloatSubtractAlt
}
