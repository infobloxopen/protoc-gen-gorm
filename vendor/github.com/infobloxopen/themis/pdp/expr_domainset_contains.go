package pdp

import "fmt"

type functionSetOfDomainsContains struct {
	set   Expression
	value Expression
}

func makeFunctionSetOfDomainsContains(set, value Expression) Expression {
	return functionSetOfDomainsContains{
		set:   set,
		value: value}
}

func makeFunctionSetOfDomainsContainsAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"contains\" for Set Of Domains needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionSetOfDomainsContains(args[0], args[1])
}

func (f functionSetOfDomainsContains) GetResultType() Type {
	return TypeBoolean
}

func (f functionSetOfDomainsContains) describe() string {
	return "contains"
}

func (f functionSetOfDomainsContains) Calculate(ctx *Context) (AttributeValue, error) {
	set, err := ctx.calculateSetOfDomainsExpression(f.set)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	value, err := ctx.calculateDomainExpression(f.value)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	_, ok := set.Get(value)
	return MakeBooleanValue(ok), nil
}

func functionSetOfDomainsContainsValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeSetOfDomains || args[1].GetResultType() != TypeDomain {
		return nil
	}

	return makeFunctionSetOfDomainsContainsAlt
}
