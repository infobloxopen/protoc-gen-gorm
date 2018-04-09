package pdp

import "fmt"

type functionFloatMultiply struct {
	first  Expression
	second Expression
}

func makeFunctionFloatMultiply(first, second Expression) Expression {
	return functionFloatMultiply{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatMultiplyAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"multiply\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatMultiply(args[0], args[1])
}

func (f functionFloatMultiply) GetResultType() int {
	return TypeFloat
}

func (f functionFloatMultiply) describe() string {
	return "multiply"
}

func (f functionFloatMultiply) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	res := first * second
	if err = floatErrorCheck(res); err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	return MakeFloatValue(res), nil
}

func functionFloatMultiplyValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}
	return makeFunctionFloatMultiplyAlt
}
