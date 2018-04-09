package pdp

import "fmt"

type functionBooleanNot struct {
	arg Expression
}

type functionBooleanOr struct {
	args []Expression
}

type functionBooleanAnd struct {
	args []Expression
}

func makeFunctionBooleanNot(args []Expression) Expression {
	if len(args) != 1 {
		panic(fmt.Errorf("boolean function \"not\" needs exactly one argument but got %d", len(args)))
	}

	return functionBooleanNot{arg: args[0]}
}

func (f functionBooleanNot) GetResultType() int {
	return TypeBoolean
}

func (f functionBooleanNot) describe() string {
	return "not"
}

func (f functionBooleanNot) Calculate(ctx *Context) (AttributeValue, error) {
	a, err := ctx.calculateBooleanExpression(f.arg)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	return MakeBooleanValue(!a), nil
}

func functionBooleanNotValidator(args []Expression) functionMaker {
	if len(args) != 1 || args[0].GetResultType() != TypeBoolean {
		return nil
	}

	return makeFunctionBooleanNot
}

func makeFunctionBooleanOr(args []Expression) Expression {
	if len(args) < 1 {
		panic(fmt.Errorf("boolean function \"or\" needs at least one argument but got %d", len(args)))
	}

	return functionBooleanOr{args: args}
}

func (f functionBooleanOr) GetResultType() int {
	return TypeBoolean
}

func (f functionBooleanOr) describe() string {
	return "or"
}

func (f functionBooleanOr) Calculate(ctx *Context) (AttributeValue, error) {
	for i, arg := range f.args {
		a, err := ctx.calculateBooleanExpression(arg)
		if err != nil {
			return undefinedValue, bindError(bindErrorf(err, "argument %d", i), f.describe())
		}

		if a {
			return MakeBooleanValue(true), nil
		}
	}

	return MakeBooleanValue(false), nil
}

func functionBooleanOrValidator(args []Expression) functionMaker {
	if len(args) < 1 {
		return nil
	}

	for _, arg := range args {
		if arg.GetResultType() != TypeBoolean {
			return nil
		}
	}

	return makeFunctionBooleanOr
}

func makeFunctionBooleanAnd(args []Expression) Expression {
	if len(args) < 1 {
		panic(fmt.Errorf("boolean function \"and\" needs at least one argument but got %d", len(args)))
	}

	return functionBooleanAnd{args: args}
}

func (f functionBooleanAnd) GetResultType() int {
	return TypeBoolean
}

func (f functionBooleanAnd) describe() string {
	return "and"
}

func (f functionBooleanAnd) Calculate(ctx *Context) (AttributeValue, error) {
	for i, arg := range f.args {
		a, err := ctx.calculateBooleanExpression(arg)
		if err != nil {
			return undefinedValue, bindError(bindErrorf(err, "argument %d", i), f.describe())
		}

		if !a {
			return MakeBooleanValue(false), nil
		}
	}

	return MakeBooleanValue(true), nil
}

func functionBooleanAndValidator(args []Expression) functionMaker {
	if len(args) < 1 {
		return nil
	}

	for _, arg := range args {
		if arg.GetResultType() != TypeBoolean {
			return nil
		}
	}

	return makeFunctionBooleanAnd
}
