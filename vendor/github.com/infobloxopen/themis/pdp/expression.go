package pdp

// Expression abstracts any PDP expression.
// The GetResultType method returns type of particular expression.
// The Calculate method returns calculated value for particular expression.
type Expression interface {
	GetResultType() int
	Calculate(ctx *Context) (AttributeValue, error)
}

type functionMaker func(args []Expression) Expression
type functionArgumentValidator func(args []Expression) functionMaker

// FunctionArgumentValidators maps function name to list of validators.
// For given set of arguments validator returns nil if the function
// doesn't accept the arguments or function which creates expression based
// on desired function and set of argument expressions.
var FunctionArgumentValidators = map[string][]functionArgumentValidator{
	"equal": {
		functionStringEqualValidator,
		functionIntegerEqualValidator,
		functionFloatEqualValidator,
	},
	"greater": {
		functionIntegerGreaterValidator,
		functionFloatGreaterValidator,
	},
	"add": {
		functionIntegerAddValidator,
		functionFloatAddValidator,
	},
	"subtract": {
		functionIntegerSubtractValidator,
		functionFloatSubtractValidator,
	},
	"multiply": {
		functionIntegerMultiplyValidator,
		functionFloatMultiplyValidator,
	},
	"divide": {
		functionIntegerDivideValidator,
		functionFloatDivideValidator,
	},
	"contains": {
		functionStringContainsValidator,
		functionNetworkContainsAddressValidator,
		functionSetOfStringsContainsValidator,
		functionSetOfNetworksContainsAddressValidator,
		functionSetOfDomainsContainsValidator},
	"not": {functionBooleanNotValidator},
	"or":  {functionBooleanOrValidator},
	"and": {functionBooleanAndValidator},
	"range": {
		functionIntegerRangeValidator,
		functionFloatRangeValidator,
	},
}
