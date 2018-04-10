package pdp

import "fmt"

type functionIntegerAdd struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerAdd(first, second Expression) Expression {
	return functionIntegerAdd{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerAddAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"add\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerAdd(args[0], args[1])
}

func (f functionIntegerAdd) GetResultType() int {
	return TypeInteger
}

func (f functionIntegerAdd) describe() string {
	return "add"
}

func (f functionIntegerAdd) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeIntegerValue(first + second), nil
}

func functionIntegerAddValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerAddAlt
}
