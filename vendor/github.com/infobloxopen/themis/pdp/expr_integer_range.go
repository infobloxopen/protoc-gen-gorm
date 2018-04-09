package pdp

import "fmt"

type functionIntegerRange struct {
	min Expression
	max Expression
	val Expression
}

func makeFunctionIntegerRange(min, max, val Expression) Expression {
	return functionIntegerRange{
		min: min,
		max: max,
		val: val,
	}

}

func makeFunctionIntegerRangeAlt(args []Expression) Expression {
	if len(args) != 3 {
		panic(fmt.Errorf("function \"Range\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerRange(args[0], args[1], args[2])
}

func (f functionIntegerRange) GetResultType() int {
	return TypeString
}

func (f functionIntegerRange) describe() string {
	return "range"
}

func (f functionIntegerRange) Calculate(ctx *Context) (AttributeValue, error) {
	min, err := ctx.calculateIntegerExpression(f.min)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "min argument"), f.describe())
	}

	max, err := ctx.calculateIntegerExpression(f.max)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "max argument"), f.describe())
	}

	val, err := ctx.calculateIntegerExpression(f.val)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "val argument"), f.describe())
	}

	switch {
	case val < min:
		return MakeStringValue("Below"), nil
	case max < val:
		return MakeStringValue("Above"), nil
	}
	return MakeStringValue("Within"), nil
}

func functionIntegerRangeValidator(args []Expression) functionMaker {
	if len(args) != 3 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger || args[2].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerRangeAlt
}
