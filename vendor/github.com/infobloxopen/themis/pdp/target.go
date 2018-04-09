package pdp

// Match represents match expression. Specific kind of boolean expression which
// can have two arguments. One of arguments should be immediate value and other
// should be attribute designator.
type Match struct {
	m Expression
}

// AllOf groups match expressions into boolean expression which result is true
// when all of child match expressions are true.
type AllOf struct {
	m []Match
}

// AnyOf groups AllOf expressions into boolean expression which result is true
// when at least one of child AllOf expressions is true.
type AnyOf struct {
	a []AllOf
}

// Target represents target expression for policy set, policy and rule.
// It gathers set of AnyOf expressions and matches to the request when all
// of child AnyOf expressions are true.
type Target struct {
	a []AnyOf
}

// MakeMatch creates instance of match expression.
func MakeMatch(e Expression) Match {
	return Match{m: e}
}

func (m Match) describe() string {
	return "match"
}

func (m Match) calculate(ctx *Context) (bool, error) {
	v, err := ctx.calculateBooleanExpression(m.m)
	if err != nil {
		return false, bindError(err, m.describe())
	}

	return v, nil
}

// MakeAllOf creates instance of AllOf expression.
func MakeAllOf() AllOf {
	return AllOf{m: []Match{}}
}

func (a AllOf) describe() string {
	return "all"
}

func (a AllOf) calculate(ctx *Context) (bool, error) {
	for _, e := range a.m {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, bindError(err, a.describe())
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

// Append adds match expression to the end of list of child match expressions.
func (a *AllOf) Append(item Match) {
	a.m = append(a.m, item)
}

// MakeAnyOf creates instance of AnyOf expressions.
func MakeAnyOf() AnyOf {
	return AnyOf{a: []AllOf{}}
}

func (a AnyOf) describe() string {
	return "any"
}

func (a AnyOf) calculate(ctx *Context) (bool, error) {
	for _, e := range a.a {
		v, err := e.calculate(ctx)
		if err != nil {
			return false, bindError(err, a.describe())
		}

		if v {
			return true, nil
		}
	}

	return false, nil
}

// Append adds AllOf expression to the end of list of child AllOf expressions.
func (a *AnyOf) Append(item AllOf) {
	a.a = append(a.a, item)
}

// MakeTarget creates instance of target.
func MakeTarget() Target {
	return Target{a: []AnyOf{}}
}

func (t Target) describe() string {
	return "target"
}

func (t Target) calculate(ctx *Context) (bool, boundError) {
	for _, e := range t.a {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, bindError(err, t.describe())
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

// Append adds AnyOf expression to the end of list of child AnyOf expressions.
func (t *Target) Append(item AnyOf) {
	t.a = append(t.a, item)
}

func makeMatchStatus(err boundError, effect int) Response {
	if effect == EffectDeny {
		return Response{EffectIndeterminateD, err, nil}
	}

	return Response{EffectIndeterminateP, err, nil}
}

func combineEffectAndStatus(err boundError, r Response) Response {
	if r.status != nil {
		err = newMultiError([]error{err, r.status})
	}

	if r.Effect == EffectNotApplicable {
		return Response{EffectNotApplicable, err, nil}
	}

	if r.Effect == EffectDeny || r.Effect == EffectIndeterminateD {
		return Response{EffectIndeterminateD, err, nil}
	}

	if r.Effect == EffectPermit || r.Effect == EffectIndeterminateP {
		return Response{EffectIndeterminateP, err, nil}
	}

	return Response{EffectIndeterminateDP, err, nil}
}

// TargetCompatibleArgument* identify expressions which supported as
// arguments of target compatible exporessions.
const (
	// TargetCompatibleArgumentAttributeValue stands for AttributeValue
	// expression.
	TargetCompatibleArgumentAttributeValue = iota
	// TargetCompatibleArgumentAttributeDesignator is AttributeDesignator
	// expression.
	TargetCompatibleArgumentAttributeDesignator
)

// CheckExpressionAsTargetArgument checks if given expression can be used
// as target argument. It returns expression kind and flag if the check
// is passed.
func CheckExpressionAsTargetArgument(e Expression) (int, bool) {
	switch e.(type) {
	case AttributeValue:
		return TargetCompatibleArgumentAttributeValue, true

	case AttributeDesignator:
		return TargetCompatibleArgumentAttributeDesignator, true
	}

	return 0, false
}

type twoArgumentsFunctionType func(first, second Expression) Expression

// TargetCompatibleExpressions maps name of expression and types of its
// arguments to particular expression maker.
var TargetCompatibleExpressions = map[string]map[int]map[int]twoArgumentsFunctionType{
	"equal": {
		TypeString: {
			TypeString: makeFunctionStringEqual},
		TypeInteger: {
			TypeInteger: makeFunctionIntegerEqual},
		TypeFloat: {
			TypeFloat: makeFunctionFloatEqual}},
	"greater": {
		TypeInteger: {
			TypeInteger: makeFunctionIntegerGreater},
		TypeFloat: {
			TypeFloat: makeFunctionFloatGreater}},
	"contains": {
		TypeString: {
			TypeString: makeFunctionStringContains},
		TypeAddress: {
			TypeNetwork: makeFunctionNetworkAddressContainedByNetwork},
		TypeNetwork: {
			TypeAddress: makeFunctionNetworkContainsAddress},
		TypeSetOfStrings: {
			TypeString: makeFunctionSetOfStringsContains},
		TypeSetOfNetworks: {
			TypeAddress: makeFunctionSetOfNetworksContainsAddress},
		TypeSetOfDomains: {
			TypeDomain: makeFunctionSetOfDomainsContains}}}
