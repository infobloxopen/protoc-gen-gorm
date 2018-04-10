package pdp

import "fmt"

type functionIntegerGreater struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerGreater(first, second Expression) Expression {
	return functionIntegerGreater{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerGreaterAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"greater\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerGreater(args[0], args[1])
}

func (f functionIntegerGreater) GetResultType() int {
	return TypeBoolean
}

func (f functionIntegerGreater) describe() string {
	return "greater"
}

func (f functionIntegerGreater) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeBooleanValue(first > second), nil
}

func functionIntegerGreaterValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerGreaterAlt
}
