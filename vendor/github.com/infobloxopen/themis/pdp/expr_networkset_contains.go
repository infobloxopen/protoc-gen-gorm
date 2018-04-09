package pdp

import "fmt"

type functionSetOfNetworksContainsAddress struct {
	set   Expression
	value Expression
}

func makeFunctionSetOfNetworksContainsAddress(set, value Expression) Expression {
	return functionSetOfNetworksContainsAddress{
		set:   set,
		value: value}
}

func makeFunctionSetOfNetworksContainsAddressAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"contains\" for Set Of Networks (Address) needs exactly two arguments but got %d",
			len(args)))
	}

	return makeFunctionSetOfNetworksContainsAddress(args[0], args[1])
}

func (f functionSetOfNetworksContainsAddress) GetResultType() int {
	return TypeBoolean
}

func (f functionSetOfNetworksContainsAddress) describe() string {
	return "contains"
}

// Calculate implements Expression interface and returns calculated value
func (f functionSetOfNetworksContainsAddress) Calculate(ctx *Context) (AttributeValue, error) {
	set, err := ctx.calculateSetOfNetworksExpression(f.set)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	a, err := ctx.calculateAddressExpression(f.value)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	_, ok := set.GetByIP(a)
	return MakeBooleanValue(ok), nil
}

func functionSetOfNetworksContainsAddressValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeSetOfNetworks || args[1].GetResultType() != TypeAddress {
		return nil
	}

	return makeFunctionSetOfNetworksContainsAddressAlt
}
