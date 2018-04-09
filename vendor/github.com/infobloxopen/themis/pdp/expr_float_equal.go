package pdp

import "fmt"

type functionFloatEqual struct {
	first  Expression
	second Expression
}

func makeFunctionFloatEqual(first, second Expression) Expression {
	return functionFloatEqual{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatEqual(args[0], args[1])
}

func (f functionFloatEqual) GetResultType() int {
	return TypeBoolean
}

func (f functionFloatEqual) describe() string {
	return "equal"
}

func (f functionFloatEqual) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeBooleanValue(first == second), nil
}

func functionFloatEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}
	return makeFunctionFloatEqualAlt
}
