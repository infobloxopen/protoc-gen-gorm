package pdp

import "fmt"

type functionFloatGreater struct {
	first  Expression
	second Expression
}

func makeFunctionFloatGreater(first, second Expression) Expression {
	return functionFloatGreater{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatGreaterAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"greater\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatGreater(args[0], args[1])
}

func (f functionFloatGreater) GetResultType() int {
	return TypeBoolean
}

func (f functionFloatGreater) describe() string {
	return "greater"
}

func (f functionFloatGreater) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeBooleanValue(first > second), nil
}

func functionFloatGreaterValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}
	return makeFunctionFloatGreaterAlt
}
