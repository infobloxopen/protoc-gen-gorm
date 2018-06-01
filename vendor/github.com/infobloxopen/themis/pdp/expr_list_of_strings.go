package pdp

import "fmt"

type functionListOfStrings struct {
	e Expression
}

func makeFunctionListOfStrings(e Expression) Expression {
	return functionListOfStrings{
		e: e,
	}
}

func makeFunctionListOfStringsAlt(args []Expression) Expression {
	if len(args) != 1 {
		panic(fmt.Errorf("function \"list of strings\" for Flags needs exactly one argument but got %d", len(args)))
	}

	return makeFunctionListOfStrings(args[0])
}

func (f functionListOfStrings) GetResultType() Type {
	return TypeListOfStrings
}

func (f functionListOfStrings) describe() string {
	return "list of strings"
}

// Calculate implements Expression interface and returns calculated value
func (f functionListOfStrings) Calculate(ctx *Context) (AttributeValue, error) {
	t := f.e.GetResultType()

	switch t {
	case TypeSetOfStrings:
		v, err := ctx.calculateSetOfStringsExpression(f.e)
		if err != nil {
			return UndefinedValue, bindError(err, f.describe())
		}

		return MakeListOfStringsValue(sortSetOfStrings(v)), nil

	case TypeListOfStrings:
		v, err := f.e.Calculate(ctx)
		if err != nil {
			return UndefinedValue, bindError(err, f.describe())
		}

		return v, nil
	}

	if t, ok := t.(*FlagsType); ok {
		var n uint64
		switch t.c {
		case 8:
			n8, err := ctx.calculateFlags8Expression(f.e)
			if err != nil {
				return UndefinedValue, bindError(err, f.describe())
			}

			n = uint64(n8)

		case 16:
			n16, err := ctx.calculateFlags16Expression(f.e)
			if err != nil {
				return UndefinedValue, bindError(err, f.describe())
			}

			n = uint64(n16)

		case 32:
			n32, err := ctx.calculateFlags32Expression(f.e)
			if err != nil {
				return UndefinedValue, bindError(err, f.describe())
			}

			n = uint64(n32)

		case 64:
			n64, err := ctx.calculateFlags64Expression(f.e)
			if err != nil {
				return UndefinedValue, bindError(err, f.describe())
			}

			n = n64
		}

		flags := make([]string, 0, len(t.b))
		for i := 0; i < len(t.b); i++ {
			if n&(1<<uint(i)) != 0 {
				flags = append(flags, t.b[i])
			}
		}

		return MakeListOfStringsValue(flags), nil
	}

	return UndefinedValue, bindError(newListOfStringsTypeError(t), f.describe())
}

func functionListOfStringsValidator(args []Expression) functionMaker {
	if len(args) == 1 {
		t := args[0].GetResultType()

		if t == TypeListOfStrings || t == TypeSetOfStrings {
			return makeFunctionListOfStringsAlt
		}

		if _, ok := t.(*FlagsType); ok {
			return makeFunctionListOfStringsAlt
		}
	}

	return nil
}
