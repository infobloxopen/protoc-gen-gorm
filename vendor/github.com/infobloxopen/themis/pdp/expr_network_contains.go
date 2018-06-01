package pdp

type functionNetworkContainsAddress struct {
	network Expression
	address Expression
}

func makeFunctionNetworkContainsAddress(network, address Expression) Expression {
	return functionNetworkContainsAddress{
		network: network,
		address: address}
}

func makeFunctionNetworkAddressContainedByNetwork(address, network Expression) Expression {
	return makeFunctionNetworkContainsAddress(network, address)
}

func makeFunctionNetworkContainsAddressAlt(args []Expression) Expression {
	if len(args) != 2 {
		return nil
	}

	return makeFunctionNetworkContainsAddress(args[0], args[1])
}

func (f functionNetworkContainsAddress) GetResultType() Type {
	return TypeBoolean
}

func (f functionNetworkContainsAddress) describe() string {
	return "contains"
}

// Calculate implements Expression interface and returns calculated value
func (f functionNetworkContainsAddress) Calculate(ctx *Context) (AttributeValue, error) {
	n, err := ctx.calculateNetworkExpression(f.network)
	if err != nil {
		return UndefinedValue, bindError(err, f.describe())
	}

	a, err := ctx.calculateAddressExpression(f.address)
	if err != nil {
		return UndefinedValue, bindError(err, f.describe())
	}

	return MakeBooleanValue(n.Contains(a)), nil
}

func functionNetworkContainsAddressValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeNetwork || args[1].GetResultType() != TypeAddress {
		return nil
	}

	return makeFunctionNetworkContainsAddressAlt
}
